package chess

import (
	"errors"
	"fmt"
)

// notationMove represents a potential move from a source square to a destination square.
type notationMove struct {
	Dest *Square // The destination square of the move.
	Src  *Square // The source square of the move.
}

// FEN applies the move to a board state represented by a FEN string
// and returns the resulting FEN string.
func (nm *notationMove) FEN(fen string) (string, error) {
	// create a board from the FEN
	board, err := loadBoard(fen)
	if err != nil {
		return "", err
	}

	srcSquare := board.GetSquare(nm.Src.File, nm.Src.Rank)
	destSquare := board.GetSquare(nm.Dest.File, nm.Dest.Rank)

	if srcSquare == nil || destSquare == nil || srcSquare.Piece == nil {
		return "", errors.New("invalid source or destination square for the given FEN")
	}

	// Use a game client to properly handle game state transitions like turn changes.
	client, err := CreateAlgebraicGameClientFromFEN(fen)
	if err != nil {
		return "", err
	}

	// Find the algebraic notation for the move to pass to the client.
	// This is necessary because client.Move expects algebraic notation,
	// and we need to ensure the move is valid in the context of the FEN.
	status, err := client.Status(true) // Force update to get valid moves
	if err != nil {
		return "", err
	}

	var moveNotation string
	for notation, move := range status.NotatedMoves {
		// Compare source and destination squares by their names (e.g., "e2", "e4")
		if move.Src.name() == nm.Src.name() && move.Dest.name() == nm.Dest.name() {
			moveNotation = notation
			break
		}
	}

	if moveNotation == "" {
		return "", fmt.Errorf("move from %s to %s is not valid for the given FEN", nm.Src.name(), nm.Dest.name())
	}

	_, err = client.Move(moveNotation, false)
	if err != nil {
		return "", fmt.Errorf("failed to apply move %s: %w", moveNotation, err)
	}

	return client.FEN(), nil
}
