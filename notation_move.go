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
	// create a brd from the FEN
	brd, err := loadBoard(fen)
	if err != nil {
		return "", err
	}

	src := brd.GetSquare(nm.Src.File, nm.Src.Rank)
	dst := brd.GetSquare(nm.Dest.File, nm.Dest.Rank)

	if src == nil || dst == nil || src.Piece == nil {
		return "", errors.New("invalid source or destination square for the given FEN")
	}

	// Use a game ac to properly handle game state transitions like turn changes.
	ac, err := CreateAlgebraicGameClientFromFEN(fen)
	if err != nil {
		return "", err
	}

	// Find the algebraic notation for the move to pass to the client.
	// This is necessary because client.Move expects algebraic notation,
	// and we need to ensure the move is valid in the context of the FEN.
	sts, err := ac.Status(true) // Force update to get valid moves
	if err != nil {
		return "", err
	}

	var not string
	for n, mv := range sts.NotatedMoves {
		// Compare source and destination squares by their names (e.g., "e2", "e4")
		if mv.Src.name() == nm.Src.name() && mv.Dest.name() == nm.Dest.name() {
			not = n
			break
		}
	}

	if not == "" {
		return "", fmt.Errorf("move from %s to %s is not valid for the given FEN", nm.Src.name(), nm.Dest.name())
	}

	_, err = ac.Move(not)
	if err != nil {
		return "", fmt.Errorf("failed to apply move %s: %w", not, err)
	}

	return ac.FEN(), nil
}
