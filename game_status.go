package chess

// gameStatus represents the state of the game at a certain point.
type gameStatus struct {
	Board        *Board                  // The current board state.
	IsCheck      bool                    // True if the current player is in check.
	IsCheckmate  bool                    // True if the current player is in checkmate.
	IsRepetition bool                    // True if the current board state is a result of repetition.
	IsStalemate  bool                    // True if the game is a stalemate.
	NotatedMoves map[string]notationMove // A map of all valid moves in algebraic notation.
}

// Side returns the side of the player who made the last move.
// If no moves have been made, it defaults to sideWhite.
func (s *gameStatus) Side() Side {
	if s.Board.LastMovedPiece == nil {
		return sideWhite
	}
	return s.Board.LastMovedPiece.Side
}
