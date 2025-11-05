package chess

import "errors"

// pieceValidator is an internal helper to determine the potential moves for a single piece
// based on its type and the rules of chess, without considering checks or other game-state constraints.
type pieceValidator struct {
	board            *Board
	allowBackward    bool
	allowDiagonal    bool
	allowForward     bool
	allowHorizontal  bool
	repeat           int
	pieceType        pieceType
	specialValidator func(v *pieceValidator, pm *potentialMoves)
}

// CreatePieceValidator initializes a new validator for a specific piece type on a given board.
// It configures movement rules (e.g., diagonal, horizontal) and repetition counts based on the piece type.
func CreatePieceValidator(pt pieceType, b *Board) *pieceValidator {
	v := &pieceValidator{
		board:     b,
		repeat:    1,
		pieceType: pt,
	}

	switch pt {
	case pieceBishop:
		v.allowDiagonal = true
		v.repeat = 8
	case pieceKing:
		v.allowBackward = true
		v.allowDiagonal = true
		v.allowForward = true
		v.allowHorizontal = true
		v.repeat = 1
	case pieceKnight:
		v.repeat = 1
		v.specialValidator = knightSpecial
	case piecePawn:
		v.allowForward = true
		v.repeat = 1
		v.specialValidator = pawnSpecial
	case pieceQueen:
		v.allowBackward = true
		v.allowDiagonal = true
		v.allowForward = true
		v.allowHorizontal = true
		v.repeat = 8
	case pieceRook:
		v.allowBackward = true
		v.allowForward = true
		v.allowHorizontal = true
		v.repeat = 8
	}

	return v
}

// Check calculates all potential destination squares for a piece at a given origin square.
// It does not validate whether a move would result in the king being in check.
// It returns a slice of valid destination squares or an error if the origin is invalid.
func (v *pieceValidator) Check(origin *Square) ([]*Square, error) {
	if origin == nil || origin.Piece == nil || origin.Piece.Type != v.pieceType {
		return nil, errors.New("piece is invalid")
	}

	ctx := &potentialMoves{
		origin:             origin,
		destinationSquares: []*Square{},
		piece:              origin.Piece,
	}

	findMoveOptions := func(direction neighbor) {
		current := v.board.getNeighborSquare(ctx.origin, direction)
		steps := 0

		for current != nil && steps < v.repeat {
			block := current.Piece != nil && (ctx.piece.Type == piecePawn || current.Piece.Side == ctx.piece.Side)
			capture := current.Piece != nil && !block

			if !block {
				ctx.destinationSquares = append(ctx.destinationSquares, current)
			}

			if capture || block {
				break
			}

			current = v.board.getNeighborSquare(current, direction)
			steps++
		}
	}

	if v.allowForward {
		forward := NeighborBelow
		if ctx.piece.Side == sideWhite {
			forward = NeighborAbove
		}
		findMoveOptions(forward)
	}

	if v.allowBackward {
		backward := NeighborAbove
		if ctx.piece.Side == sideWhite {
			backward = NeighborBelow
		}
		findMoveOptions(backward)
	}

	if v.allowHorizontal {
		findMoveOptions(NeighborLeft)
		findMoveOptions(NeighborRight)
	}

	if v.allowDiagonal {
		findMoveOptions(NeighborAboveLeft)
		findMoveOptions(NeighborAboveRight)
		findMoveOptions(NeighborBelowLeft)
		findMoveOptions(NeighborBelowRight)
	}

	if v.specialValidator != nil {
		v.specialValidator(v, ctx)
	}

	return ctx.destinationSquares, nil
}

// knightSpecial implements the special L-shaped move validation for a Knight.
// It calculates the eight potential squares a knight can jump to.
func knightSpecial(v *pieceValidator, pm *potentialMoves) {
	cndts := []*Square{}

	aL := v.board.getNeighborSquare(pm.origin, NeighborAboveLeft)
	aR := v.board.getNeighborSquare(pm.origin, NeighborAboveRight)
	bL := v.board.getNeighborSquare(pm.origin, NeighborBelowLeft)
	bR := v.board.getNeighborSquare(pm.origin, NeighborBelowRight)

	if aL != nil {
		cndts = append(cndts,
			v.board.getNeighborSquare(aL, NeighborAbove),
			v.board.getNeighborSquare(aL, NeighborLeft),
		)
	}

	if aR != nil {
		cndts = append(cndts,
			v.board.getNeighborSquare(aR, NeighborAbove),
			v.board.getNeighborSquare(aR, NeighborRight),
		)
	}

	if bL != nil {
		cndts = append(cndts,
			v.board.getNeighborSquare(bL, NeighborBelow),
			v.board.getNeighborSquare(bL, NeighborLeft),
		)
	}

	if bR != nil {
		cndts = append(cndts,
			v.board.getNeighborSquare(bR, NeighborBelow),
			v.board.getNeighborSquare(bR, NeighborRight),
		)
	}

	for _, sq := range cndts {
		if sq == nil {
			continue
		}

		if sq.Piece == nil || sq.Piece.Side != pm.piece.Side {
			pm.destinationSquares = append(pm.destinationSquares, sq)
		}
	}
}

// pawnSpecial implements the special move validation for a Pawn.
// This includes:
// - Standard single-square forward move.
// - Initial two-square forward move.
// - Diagonal captures.
// - En passant captures.
func pawnSpecial(v *pieceValidator, pm *potentialMoves) {
	dL := v.board.getNeighborSquare(pm.origin, NeighborBelowLeft)
	dR := v.board.getNeighborSquare(pm.origin, NeighborBelowRight)
	if pm.piece.Side == sideWhite {
		dL = v.board.getNeighborSquare(pm.origin, NeighborAboveLeft)
		dR = v.board.getNeighborSquare(pm.origin, NeighborAboveRight)
	}

	// check for capture Left and capture Right
	for _, sq := range []*Square{dL, dR} {
		if sq == nil {
			continue
		}

		if sq.Piece != nil && sq.Piece.Side != pm.piece.Side {
			pm.destinationSquares = append(pm.destinationSquares, sq)
		}
	}

	// check for initial 2-square forward move
	if pm.piece.MoveCount == 0 && len(pm.destinationSquares) > 0 {
		ff := pm.destinationSquares[0]
		if ff != nil && ff.Piece == nil && (pm.origin.Rank == 7 || pm.origin.Rank == 2) {
			var sf *Square
			fD := NeighborBelow

			if pm.piece.Side == sideWhite {
				fD = NeighborAbove
			}

			sf = v.board.getNeighborSquare(ff, fD)
			if sf != nil && sf.Piece == nil {
				pm.destinationSquares = append(pm.destinationSquares, sf)
			}
		}
	}

	// determine rank for en-passant
	rnk := 5
	if pm.piece.Side == sideBlack {
		rnk = 4
	}

	if pm.origin.Rank != rnk {
		return
	}

	left := v.board.getNeighborSquare(pm.origin, NeighborLeft)
	right := v.board.getNeighborSquare(pm.origin, NeighborRight)

	for _, adj := range []*Square{left, right} {
		if adj == nil || adj.Piece == nil {
			continue
		}

		if adj.Piece.Type == piecePawn &&
			adj.Piece.Side != pm.piece.Side &&
			adj.Piece.MoveCount == 1 &&
			v.board.LastMovedPiece == adj.Piece {

			cd := NeighborBelow
			if adj.Piece.Side == sideBlack {
				cd = NeighborAbove
			}

			ct := v.board.getNeighborSquare(adj, cd)
			if ct != nil {
				pm.destinationSquares = append(pm.destinationSquares, ct)
			}
		}
	}
}
