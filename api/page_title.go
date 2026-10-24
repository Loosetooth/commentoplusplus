package main

import (
	"net/url"
	"os"
)

func pageTitleUpdate(domain string, path string) (string, error) {
	u := url.URL{Scheme: "http", Host: domain, Path: path}
	if os.Getenv("SSL") == "true" {
		u.Scheme = "https"
	}
	// Building the URL from parts (instead of concatenating strings) ensures
	// that a client-supplied path like "@internal-host/" can't change the host.
	title, err := htmlTitleGet(u.String())
	if err != nil {
		// This could fail due to a variety of reasons that we can't control such
		// as the user's URL 404 or something, so let's not pollute the error log
		// with messages. Just use a sane title. Maybe we'll have the ability to
		// retry later.
		logger.Errorf("%v", err)
		title = domain
	}

	statement := `
		UPDATE pages
		SET title = $3
		WHERE canon($1) LIKE canon(domain) AND path = $2;
	`
	_, err = db.Exec(statement, domain, path, title)
	if err != nil {
		logger.Errorf("cannot update pages table with title: %v", err)
		return "", err
	}

	return title, nil
}
