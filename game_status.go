package chess

// GameStatus represents the state of the game at a certain point in time.
type GameStatus struct {
	Game         *Game                   // The current game state.
	IsCheck      bool                    // True if the current player is in check.
	IsCheckmate  bool                    // True if the current player is in checkmate.
	IsRepetition bool                    // True if the current board state is a result of repetition.
	IsStalemate  bool                    // True if the game is a stalemate.
	NotatedMoves map[string]notationMove // A map of all valid moves in algebraic notation.
}

// Side returns the side of the player who made the last move.
// If no moves have been made, it defaults to sideWhite (unless this
// state has been overridden for the game).
func (s *GameStatus) Side() Side {
	if s.Game.Board.LastMovedPiece == nil {
		if s.Game.wf {
			return sideWhite
		}

		return sideBlack
	}

	return s.Game.Board.LastMovedPiece.Side
}
