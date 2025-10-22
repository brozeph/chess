package chess

import "testing"

func TestPawnCapture(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	result := mustMove(t, client, "exd5")

	if result.Move.CapturedPiece == nil || result.Move.CapturedPiece.Type != piecePawn {
		t.Fatalf("expected captured pawn, got %+v", result.Move.CapturedPiece)
	}
}

func TestMoveHistoryRecordsNotation(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	mustMove(t, client, "exd5")

	if len(client.game.MoveHistory) < 3 {
		t.Fatalf("expected at least 3 moves, got %d", len(client.game.MoveHistory))
	}
	if client.game.MoveHistory[2].Algebraic != "exd5" {
		t.Fatalf("expected notation exd5, got %s", client.game.MoveHistory[2].Algebraic)
	}
}

func TestCaptureHistoryAndUndo(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	capture := mustMove(t, client, "exd5")

	history := client.CaptureHistory()
	if len(history) != 1 {
		t.Fatalf("expected capture history length 1, got %d", len(history))
	}
	if history[0].Type != piecePawn {
		t.Fatalf("expected pawn capture history, got %v", history[0].Type)
	}

	capture.Undo()

	history = client.CaptureHistory()
	if len(history) != 0 {
		t.Fatalf("expected empty capture history after undo, got %d", len(history))
	}
}

func TestKnightDisambiguation(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "Nc3")
	mustMove(t, client, "Nf6")
	mustMove(t, client, "Nd5")
	mustMove(t, client, "Ng8")
	mustMove(t, client, "Nf4")
	mustMove(t, client, "Nf6")

	status := mustStatus(t, client, false)

	if _, ok := status.NotatedMoves["Nfh3"]; !ok {
		t.Fatalf("expected Nfh3 in notated moves")
	}
	if _, ok := status.NotatedMoves["Ngh3"]; !ok {
		t.Fatalf("expected Ngh3 in notated moves")
	}
}

func TestRookDisambiguationRanks(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "a4")
	mustMove(t, client, "a5")
	mustMove(t, client, "h4")
	mustMove(t, client, "h5")
	mustMove(t, client, "Ra3")
	mustMove(t, client, "Ra6")
	mustMove(t, client, "Rhh3")
	mustMove(t, client, "Rhh6")

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["Rae3"]; !ok {
		t.Fatalf("expected Rae3")
	}
	if _, ok := status.NotatedMoves["Rhe3"]; !ok {
		t.Fatalf("expected Rhe3")
	}
}

func TestRookDisambiguationFiles(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	sequence := []string{
		"a4", "a5", "h4", "h5", "Ra3", "Ra6", "Rhh3", "Rhh6",
		"Rae3", "Rh8", "Re6", "Ra8", "Rhe3", "Ra6",
	}

	for _, mv := range sequence {
		mustMove(t, client, mv)
	}

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["R6e5"]; !ok {
		t.Fatalf("expected R6e5")
	}
	if _, ok := status.NotatedMoves["R3e5"]; !ok {
		t.Fatalf("expected R3e5")
	}
}

func TestAmbiguousNotationThrows(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	sequence := []string{
		"a4", "a5", "h4", "h5", "Ra3", "Ra6",
	}
	for _, mv := range sequence {
		mustMove(t, client, mv)
	}

	if _, err := client.Move("Rh3", false); err == nil {
		t.Fatalf("expected ambiguous notation error")
	}
}

func TestInvalidNotationThrows(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	if _, err := client.Move("h6", false); err == nil {
		t.Fatalf("expected move error for h6")
	}
	if _, err := client.Move("z9", false); err == nil {
		t.Fatalf("expected move error for z9")
	}
}

func TestVerboseNotationParses(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	res := mustMove(t, client, "Nb1c3")
	if res.Move.PostSquare.File != 'c' || res.Move.PostSquare.Rank != 3 {
		t.Fatalf("expected knight on c3")
	}
	if res.Move.PostSquare.Piece == nil || res.Move.PostSquare.Piece.Type != pieceKnight {
		t.Fatalf("expected knight piece")
	}
}
