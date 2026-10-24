package main

import (
	"net/http"
)

func commentEdit(commentHex string, markdown string, state string, url string) (string, error) {
	if commentHex == "" {
		return "", errorMissingField
	}

	html := markdownToHtml(markdown)

	statement := `
		UPDATE comments
		SET markdown = $2, html = $3, state = $4
		WHERE commentHex=$1;
	`
	_, err := db.Exec(statement, commentHex, markdown, html, state)

	if err != nil {
		// TODO: make sure this is the error is actually non-existant commentHex
		return "", errorNoSuchComment
	}

	hub.broadcast <- []byte(url)

	return html, nil
}

func commentEditHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"commenterToken"`
		CommentHex     *string `json:"commentHex"`
		Markdown       *string `json:"markdown"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	domain, path, err := commentDomainPathGet(*x.CommentHex)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	c, err := commenterGetByCommenterToken(*x.CommenterToken)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	cm, err := commentGetByCommentHex(*x.CommentHex)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	if cm.CommenterHex != c.CommenterHex {
		bodyMarshal(w, response{"success": false, "message": errorNotAuthorised.Error()})
		return
	}

	d, err := domainGet(domain)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	if d.State == "frozen" {
		bodyMarshal(w, response{"success": false, "message": errorDomainFrozen.Error()})
		return
	}

	p, err := pageGet(domain, path)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	if p.IsLocked {
		bodyMarshal(w, response{"success": false, "message": errorThreadLocked.Error()})
		return
	}

	isModerator, err := isDomainModerator(domain, c.Email)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	// Editing must not be a way around moderation, so apply the same checks as
	// commentNewHandler does to new comments.
	state := cm.State
	if !isModerator {
		if d.RequireModeration {
			state = "unapproved"
		} else if d.AutoSpamFilter && (isSpam(domain, getIp(r), getUserAgent(r), c.Name, c.Email, c.Link, *x.Markdown) || isToxic(*x.Markdown)) {
			state = "flagged"
		}
	}

	html, err := commentEdit(*x.CommentHex, *x.Markdown, state, domain+path)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshal(w, response{"success": true, "html": html})
}
