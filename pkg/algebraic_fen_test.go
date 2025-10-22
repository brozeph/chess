package chess

import "testing"

func TestGetFEN(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	if got := client.FEN(); got != "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR" {
		t.Fatalf("unexpected FEN %s", got)
	}
}

func TestFromFENRespectsSideToMove(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1"
	client, err := CreateAlgebraicGameClientFromFEN(fen, AlgebraicClientOptions{})
	if err != nil {
		t.Fatalf("fromFEN failed: %v", err)
	}

	expectedFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
	if got := client.FEN(); got != expectedFEN {
		t.Fatalf("board mismatch: got %s want %s", got, expectedFEN)
	}

	if _, err := client.Move("e4", false); err == nil {
		t.Fatalf("expected white move to fail")
	}

	res := mustMove(t, client, "e5")
	if res.Move.PostSquare.File != 'e' || res.Move.PostSquare.Rank != 5 {
		t.Fatalf("expected move to e5")
	}
	if res.Move.PostSquare.Piece == nil || res.Move.PostSquare.Piece.Side != sideBlack {
		t.Fatalf("expected black piece")
	}
}
