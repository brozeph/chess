package chess

import "testing"

func mustMove(t *testing.T, client *AlgebraicGameClient, notation string) *moveResult {
	t.Helper()
	res, err := client.Move(notation, false)
	if err != nil {
		t.Fatalf("move %s failed: %v", notation, err)
	}
	return res
}

func mustStatus(t *testing.T, client *AlgebraicGameClient, force bool) *clientStatus {
	t.Helper()
	sts, err := client.Status(force)
	if err != nil {
		t.Fatalf("getStatus failed: %v", err)
	}
	return sts
}
