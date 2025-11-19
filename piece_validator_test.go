package chess

import "testing"

func TestBoardMoveDoesNotFlagNonCastlingKingMoves(t *testing.T) {
	fen := "r2q3r/p1p1bk1p/b1P2Np1/1p2Qp2/8/2N3P1/PPPP1PBP/R1B2RK1 b - - 14 1"
	board, err := loadBoard(fen)
	if err != nil {
		t.Fatalf("failed to load fen: %v", err)
	}

	src := board.GetSquare('f', 7)
	dst := board.GetSquare('g', 7)
	if src == nil || dst == nil || src.Piece == nil {
		t.Fatalf("missing king squares for test")
	}

	res, err := board.Move(src, dst, true, "")
	if err != nil {
		t.Fatalf("simulate move failed: %v", err)
	}
	defer res.Undo()

	if res.Move.Castle {
		t.Fatalf("king move f7->g7 should not be treated as castling")
	}
}
