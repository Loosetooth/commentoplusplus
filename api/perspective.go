package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/Jeffail/gabs"
)

func isToxic(markdown string) bool {

	perspectiveKey := os.Getenv("PERSPECTIVE_KEY")

	if perspectiveKey == "" {
		return false
	}

	pl := os.Getenv("PERSPECTIVE_LIMIT")
	perspectiveLimit, err := strconv.ParseFloat(pl, 32)

	if err != nil {
		logger.Errorf("error: Perspective limit not ok: %v", err)
		return true
	}

	perspectiveLang := os.Getenv("PERSPECTIVE_LANGUAGE")

	url := "https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze"
	method := "POST"

	payload, err := json.Marshal(map[string]interface{}{
		"comment":             map[string]string{"text": markdown},
		"languages":           []string{perspectiveLang},
		"requestedAttributes": map[string]interface{}{"TOXICITY": map[string]interface{}{}},
	})
	if err != nil {
		logger.Errorf("error: cannot validate comment using Perspective: %v", err)
		return true
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))

	if err != nil {
		logger.Errorf("error: cannot validate comment using Perspective: %v", err)
		return true
	}
	req.Header.Add("Content-Type", "application/json")
	// Keep the API key out of the URL so that it doesn't end up in error logs.
	req.Header.Add("X-Goog-Api-Key", perspectiveKey)

	res, err := client.Do(req)
	if err != nil {
		logger.Errorf("error: cannot validate comment using Perspective: %v", err)
		return true
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Errorf("error: cannot validate comment using Perspective: %v", err)
		return true
	}

	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		logger.Errorf("error: cannot parsePerspective result", err)
		return true
	}

	var value float64
	var ok bool
	var resp bool

	value, ok = jsonParsed.Path("attributeScores.TOXICITY.summaryScore.value").Data().(float64)

	if !ok {
		logger.Errorf("error: no valid Perspective value")
		return true
	}

	if value > perspectiveLimit {
		resp = true
	} else {
		resp = false
	}

	return resp
}
