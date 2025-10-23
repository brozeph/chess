package chess

import "testing"

func TestCreateInitialStatus(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	status := mustStatus(t, client, false)

	if status.IsCheck {
		t.Fatalf("expected IsCheck false")
	}
	if status.IsCheckmate {
		t.Fatalf("expected IsCheckmate false")
	}
	if status.IsRepetition {
		t.Fatalf("expected IsRepetition false")
	}
	if status.IsStalemate {
		t.Fatalf("expected IsStalemate false")
	}
	if got := len(status.NotatedMoves); got != 20 {
		t.Fatalf("expected 20 notated moves, got %d", got)
	}
}

func TestStatusAfterMoves(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "b4")
	mustMove(t, client, "e6")

	status := mustStatus(t, client, false)

	if status.IsCheck || status.IsCheckmate || status.IsRepetition || status.IsStalemate {
		t.Fatalf("unexpected flags %#v", status)
	}
	if got := len(status.NotatedMoves); got != 21 {
		t.Fatalf("expected 21 notated moves, got %d", got)
	}
}

func TestIssue71UndoRestoresStatus(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "c5").Undo()

	status := mustStatus(t, client, false)
	if status.Board.LastMovedPiece == nil || status.Board.LastMovedPiece.Side != sideWhite {
		t.Fatalf("expected last moved piece to be white")
	}
	if _, ok := status.NotatedMoves["c5"]; !ok {
		t.Fatalf("expected c5 in notated moves")
	}
}

func TestIssue77UndoFirstMove(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "e4").Undo()

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["e4"]; !ok {
		t.Fatalf("expected e4 available after undo")
	}
}
