package parser

import "testing"

func TestIsNoiseUserTurn(t *testing.T) {
	if !isNoiseUserTurn("<task-notification>\nDone\n</task-notification>") {
		t.Fatal("expected task-notification noise")
	}
	if isNoiseUserTurn("real user question") {
		t.Fatal("expected real text")
	}
}
