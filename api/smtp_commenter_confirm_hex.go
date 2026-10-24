package main

import (
	"bytes"
	"os"
)

type commenterConfirmHexPlugs struct {
	Origin     string
	ConfirmHex string
}

func smtpCommenterConfirmHex(to string, toName string, confirmHex string) error {
	var body bytes.Buffer
	templates["commenter-confirm-hex"].Execute(&body, &commenterConfirmHexPlugs{Origin: os.Getenv("ORIGIN"), ConfirmHex: confirmHex})

	err := smtpSendMail(to, toName, "", "Please confirm your email address", body.String())
	if err != nil {
		logger.Errorf("cannot send confirmation email: %v", err)
		return errorCannotSendEmail
	}

	return nil
}
