package chess

import "testing"

const strtFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

func TestGetFEN(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	exp := strtFEN
	if got := client.FEN(); got != exp {
		t.Fatalf("mismatch: \n got: %s\nwant: %s", got, exp)
	}
}

func TestFromFENRespectsSideToMove(t *testing.T) {
	fen := strtFEN
	client, err := CreateAlgebraicGameClientFromFEN(fen, AlgebraicClientOptions{})
	if err != nil {
		t.Fatalf("fromFEN failed: %v", err)
	}

	expFEN := strtFEN
	if got := client.FEN(); got != expFEN {
		t.Fatalf("board mismatch: \n got: %s\nwant: %s", got, expFEN)
	}

	// move two spaces with white pawn
	if _, err := client.Move("e4", false); err != nil {
		t.Fatalf("expected white move to succeed")
	}

	expFEN = "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
	if got := client.FEN(); got != expFEN {
		t.Fatalf("board mismatch: \n got: %s\nwant: %s", got, expFEN)
	}

	res := mustMove(t, client, "e5")
	if res.Move.PostSquare.File != 'e' || res.Move.PostSquare.Rank != 5 {
		t.Fatalf("expected move to e5")
	}
	if res.Move.PostSquare.Piece == nil || res.Move.PostSquare.Piece.Side != sideBlack {
		t.Fatalf("expected black piece")
	}
}
