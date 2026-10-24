package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strings"
	"syscall"
	"time"
)

type response map[string]interface{}

// TODO: Add tests in utils_http_test.go

// Request bodies are small JSON objects; anything larger is rejected so that
// unauthenticated clients cannot exhaust memory.
const maxBodySize = 1 << 20

func bodyUnmarshal(r *http.Request, x interface{}) error {
	b, err := ioutil.ReadAll(http.MaxBytesReader(nil, r.Body, maxBodySize))
	if err != nil {
		logger.Errorf("cannot read POST body: %v\n", err)
		return errorInternal
	}

	if err = json.Unmarshal(b, x); err != nil {
		return errorInvalidJSONBody
	}

	xv := reflect.Indirect(reflect.ValueOf(x))
	for i := 0; i < xv.NumField(); i++ {
		if xv.Field(i).IsNil() {
			return errorMissingField
		}
	}

	return nil
}

func bodyMarshal(w http.ResponseWriter, x map[string]interface{}) error {
	resp, err := json.Marshal(x)
	if err != nil {
		w.Write([]byte(`{"success":false,"message":"Some internal error occurred"}`))
		logger.Errorf("cannot marshal response: %v\n")
		return errorInternal
	}

	w.Write(resp)
	return nil
}

func getIp(r *http.Request) string {
	ip := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// The reverse proxy appends the address it saw to whatever the client
		// sent, so only the last entry can be trusted.
		hops := strings.Split(xff, ",")
		ip = strings.TrimSpace(hops[len(hops)-1])
	}

	return ip
}

func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

var errorForbiddenAddress = errors.New("refusing to connect to a non-public address")

// publicAddressOnly is a net.Dialer Control function that only allows
// connections to public unicast addresses. It runs after DNS resolution, for
// every connection (including redirects), so it can't be bypassed with DNS
// tricks.
func publicAddressOnly(network string, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	ip := net.ParseIP(host)
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate() || cgnat.Contains(ip) {
		return errorForbiddenAddress
	}

	return nil
}

var _, cgnat, _ = net.ParseCIDR("100.64.0.0/10")

// publicHttpClient must be used for every request to a URL that is
// influenced by users (avatars, page titles, imports) so that those can't be
// used to reach services on the internal network.
var publicHttpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
			Control: publicAddressOnly,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("stopped after 5 redirects")
		}
		return nil
	},
}
