package main

import (
	"testing"
	"time"
)

func TestResetExpired(t *testing.T) {
	failTestOnError(t, setupTestEnv())

	commenterHex, _ := commenterNew("test@example.com", "Test", "undefined", "undefined", "commento", "hunter2")

	statement := `
		INSERT INTO resetHexes (resetHex, hex, entity, sendDate)
		VALUES ($1, $2, 'commenter', $3);
	`
	db.Exec(statement, "old-hex", commenterHex, time.Now().UTC().Add(-2*time.Hour))
	db.Exec(statement, "new-hex", commenterHex, time.Now().UTC())
	db.Exec(statement, "other-hex", commenterHex, time.Now().UTC())

	if _, err := reset("old-hex", "hunter3"); err != errorNoSuchResetToken {
		t.Errorf("expected errorNoSuchResetToken for an expired reset token, got: %v", err)
		return
	}

	commenterToken, _ := commenterLogin("test@example.com", "hunter2")

	if _, err := reset("new-hex", "hunter3"); err != nil {
		t.Errorf("unexpected error resetting password: %v", err)
		return
	}

	if _, err := commenterGetByCommenterToken(commenterToken); err == nil {
		t.Errorf("expected existing sessions to be invalidated by a password reset")
		return
	}

	if _, err := reset("other-hex", "hunter4"); err != errorNoSuchResetToken {
		t.Errorf("expected other reset tokens to be invalidated by a password reset, got: %v", err)
		return
	}

	if _, err := commenterLogin("test@example.com", "hunter3"); err != nil {
		t.Errorf("unexpected error logging in with the new password: %v", err)
		return
	}
}
