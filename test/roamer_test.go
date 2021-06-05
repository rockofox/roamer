package test

import (
	"testing"

	"github.com/rendon/testcli"
)

func TestRoamer(t *testing.T) {
	testcli.Run("roamer")
	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %s", testcli.Error())
	}
}
func TestFail(t *testing.T) {
	testcli.Run("roamer overview")
	if testcli.Success() {
		t.Fatalf("Expected to fail, but succeeded: %s", testcli.Error())
	}
}
