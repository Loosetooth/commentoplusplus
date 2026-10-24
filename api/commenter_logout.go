package main

import (
	"net/http"
)

func commenterLogoutHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"commenterToken"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	statement := `
		DELETE FROM commenterSessions
		WHERE commenterToken = $1;
	`
	if _, err := db.Exec(statement, *x.CommenterToken); err != nil {
		logger.Errorf("cannot delete commenter session: %v", err)
		bodyMarshal(w, response{"success": false, "message": errorInternal.Error()})
		return
	}

	bodyMarshal(w, response{"success": true})
}
