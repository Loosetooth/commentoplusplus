package main

import (
	"fmt"
	"net/http"
	"time"
)

func commenterConfirmHexNew(commenterHex string, email string, name string) error {
	confirmHex, err := randomHex(32)
	if err != nil {
		logger.Errorf("cannot generate confirmHex: %v", err)
		return errorInternal
	}

	statement := `
		INSERT INTO
		commenterConfirmHexes (confirmHex, commenterHex, sendDate)
		VALUES                ($1,         $2,           $3      );
	`
	_, err = db.Exec(statement, confirmHex, commenterHex, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert confirmHex: %v\n", err)
		return errorInternal
	}

	return smtpCommenterConfirmHex(email, name, confirmHex)
}

func commenterConfirmHex(confirmHex string) error {
	if confirmHex == "" {
		return errorMissingField
	}

	statement := `
		UPDATE commenters
		SET confirmedEmail=true
		WHERE commenterHex IN (
			SELECT commenterHex FROM commenterConfirmHexes
			WHERE confirmHex=$1
		);
	`
	res, err := db.Exec(statement, confirmHex)
	if err != nil {
		logger.Errorf("cannot mark commenter's confirmedEmail as true: %v\n", err)
		return errorInternal
	}

	count, err := res.RowsAffected()
	if err != nil {
		logger.Errorf("cannot count rows affected: %v\n", err)
		return errorInternal
	}

	if count == 0 {
		return errorNoSuchConfirmationToken
	}

	statement = `
		DELETE FROM commenterConfirmHexes
		WHERE confirmHex=$1;
	`
	_, err = db.Exec(statement, confirmHex)
	if err != nil {
		logger.Warningf("cannot remove confirmation token: %v\n", err)
		// Don't return an error because this is not critical.
	}

	return nil
}

func commenterConfirmHexHandler(w http.ResponseWriter, r *http.Request) {
	if err := commenterConfirmHex(r.FormValue("confirmHex")); err != nil {
		fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	fmt.Fprintf(w, "Your email address has been confirmed. You can now close this page and log in to comment.\n")
}
