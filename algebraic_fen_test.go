package chess

import (
	"slices"
	"testing"
)

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
	if _, err := client.Move("e4"); err != nil {
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

func TestFromFENProperMoves(t *testing.T) {
	for _, tc := range []struct {
		name               string
		fen                string
		expectedLegalMoves []string
		expectErr          bool
	}{
		{
			"should not have move option that includes d2",
			"r1bqk1nr/pppp1ppp/2n5/8/1P1pP3/5N2/PP3PPP/RNBQKB1R b KQkq - 0 6",
			[]string{
				"Ke7",
				"Kf8",
				"Nge7",
				"Nf6",
				"Nh6",
				"Nce7",
				"Nb8",
				"Na5",
				"Ne5",
				"Nxb4",
				"Qe7",
				"Qf6",
				"Qg5",
				"Qh4",
				"Rb8",
				"a5",
				"a6",
				"b5",
				"b6",
				"d3",
				"d5",
				"d6",
				"f5",
				"f6",
				"g5",
				"g6",
				"h5",
				"h6",
			},
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gc, err := CreateAlgebraicGameClientFromFEN(tc.fen)
			if (err != nil) != tc.expectErr {
				t.Fatalf("FEN() error = %v, expectErr %v", err, tc.expectErr)
			}

			lm := []string{}
			for ntn := range gc.notatedMoves {
				lm = append(lm, ntn)
			}
			slices.Sort(lm)

			for _, mv := range tc.expectedLegalMoves {
				if !slices.Contains(lm, mv) {
					t.Fatalf("CreateAlgebraicGameClientFromFEN() missing legal move %s.\n got: %v\nwant: %v", mv, lm, tc.expectedLegalMoves)
				}
			}

			if len(tc.expectedLegalMoves) != len(lm) {
				t.Fatalf("CreateAlgebraicGameClientFromFEN() extra legal moves.\n got: %v\nwant: %v", lm, tc.expectedLegalMoves)
			}
		})
	}
}
