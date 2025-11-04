package chess

import "testing"

func TestNotationMove_FEN(t *testing.T) {
	// Initial board setup for creating notationMove instances
	initialBoard := createBoard()

	testCases := []struct {
		name        string
		move        notationMove
		initialFEN  string
		expectedFEN string
		expectErr   bool
	}{
		{
			name: "White pawn e2 to e4",
			move: notationMove{
				Src:  initialBoard.GetSquare('e', 2),
				Dest: initialBoard.GetSquare('e', 4),
			},
			initialFEN:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w - - 0 1", // Simplified castling for consistency
			expectedFEN: "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b - e3 0 1",
			expectErr:   false,
		},
		{
			name: "Black pawn d7 to d5",
			move: notationMove{
				Src:  initialBoard.GetSquare('d', 7),
				Dest: initialBoard.GetSquare('d', 5),
			},
			initialFEN:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b - e3 0 1", // Simplified castling for consistency
			expectedFEN: "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w - d6 0 2",
			expectErr:   false,
		},
		{
			name: "White pawn e4 captures d5",
			move: notationMove{
				Src:  initialBoard.GetSquare('e', 4),
				Dest: initialBoard.GetSquare('d', 5),
			},
			initialFEN:  "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w - d6 0 2", // Simplified castling for consistency
			expectedFEN: "rnbqkbnr/ppp1pppp/8/3P4/8/8/PPPP1PPP/RNBQKBNR b - - 0 2",
			expectErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultFEN, err := tc.move.FEN(tc.initialFEN)
			if (err != nil) != tc.expectErr {
				t.Fatalf("FEN() error = %v, expectErr %v", err, tc.expectErr)
			}

			if resultFEN != tc.expectedFEN {
				t.Errorf("FEN() \n got: %v\nwant: %v", resultFEN, tc.expectedFEN)
			}
		})
	}
}
