package main

import (
	"net/http"
)

func ownerLogoutHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	statement := `
		DELETE FROM ownerSessions
		WHERE ownerToken = $1;
	`
	if _, err := db.Exec(statement, *x.OwnerToken); err != nil {
		logger.Errorf("cannot delete owner session: %v", err)
		bodyMarshal(w, response{"success": false, "message": errorInternal.Error()})
		return
	}

	bodyMarshal(w, response{"success": true})
}
