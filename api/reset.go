package main

import (
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func reset(resetHex string, password string) (string, error) {
	if resetHex == "" || password == "" {
		return "", errorMissingField
	}

	statement := `
		SELECT hex, entity
		FROM resetHexes
		WHERE resetHex = $1 AND sendDate::timestamptz > $2;
	`
	row := db.QueryRow(statement, resetHex, time.Now().UTC().Add(-1*time.Hour))

	var hex string
	var entity string
	if err := row.Scan(&hex, &entity); err != nil {
		// TODO: is this the only error?
		return "", errorNoSuchResetToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("cannot generate hash from password: %v\n", err)
		return "", errorInternal
	}

	if entity == "owner" {
		statement = `
			UPDATE owners SET passwordHash = $1, confirmedEmail=true
			WHERE ownerHex = $2;
		`
	} else {
		// Receiving the reset email proves ownership of the address.
		statement = `
			UPDATE commenters SET passwordHash = $1, confirmedEmail=true
			WHERE commenterHex = $2;
		`
	}

	_, err = db.Exec(statement, string(passwordHash), hex)
	if err != nil {
		logger.Errorf("cannot change %s's password: %v\n", entity, err)
		return "", errorInternal
	}

	statement = `
		DELETE FROM resetHexes
		WHERE hex = $1;
	`
	_, err = db.Exec(statement, hex)
	if err != nil {
		logger.Warningf("cannot remove resetHex: %v\n", err)
	}

	// Log out every existing session, in case the password was reset because
	// the account was compromised.
	if entity == "owner" {
		statement = `
			DELETE FROM ownerSessions
			WHERE ownerHex = $1;
		`
	} else {
		statement = `
			DELETE FROM commenterSessions
			WHERE commenterHex = $1;
		`
	}
	_, err = db.Exec(statement, hex)
	if err != nil {
		logger.Errorf("cannot remove %s's sessions: %v\n", entity, err)
		return "", errorInternal
	}

	return entity, nil
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		ResetHex *string `json:"resetHex"`
		Password *string `json:"password"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	entity, err := reset(*x.ResetHex, *x.Password)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshal(w, response{"success": true, "entity": entity})
}
