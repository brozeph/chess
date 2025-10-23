package chess

import "errors"

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

func knightSpecial(v *pieceValidator, pm *potentialMoves) {
	candidates := []*Square{}

	aboveLeft := v.board.getNeighborSquare(pm.origin, NeighborAboveLeft)
	aboveRight := v.board.getNeighborSquare(pm.origin, NeighborAboveRight)
	belowLeft := v.board.getNeighborSquare(pm.origin, NeighborBelowLeft)
	belowRight := v.board.getNeighborSquare(pm.origin, NeighborBelowRight)

	if aboveLeft != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(aboveLeft, NeighborAbove),
			v.board.getNeighborSquare(aboveLeft, NeighborLeft),
		)
	}

	if aboveRight != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(aboveRight, NeighborAbove),
			v.board.getNeighborSquare(aboveRight, NeighborRight),
		)
	}

	if belowLeft != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(belowLeft, NeighborBelow),
			v.board.getNeighborSquare(belowLeft, NeighborLeft),
		)
	}

	if belowRight != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(belowRight, NeighborBelow),
			v.board.getNeighborSquare(belowRight, NeighborRight),
		)
	}

	for _, sq := range candidates {
		if sq == nil {
			continue
		}

		if sq.Piece == nil || sq.Piece.Side != pm.piece.Side {
			pm.destinationSquares = append(pm.destinationSquares, sq)
		}
	}
}

func pawnSpecial(v *pieceValidator, pm *potentialMoves) {
	diagLeft := v.board.getNeighborSquare(pm.origin, NeighborBelowLeft)
	diagRight := v.board.getNeighborSquare(pm.origin, NeighborBelowRight)
	if pm.piece.Side == sideWhite {
		diagLeft = v.board.getNeighborSquare(pm.origin, NeighborAboveLeft)
		diagRight = v.board.getNeighborSquare(pm.origin, NeighborAboveRight)
	}

	for _, sq := range []*Square{diagLeft, diagRight} {
		if sq == nil {
			continue
		}
		if sq.Piece != nil && sq.Piece.Side != pm.piece.Side {
			pm.destinationSquares = append(pm.destinationSquares, sq)
		}
	}

	if pm.piece.MoveCount == 0 && len(pm.destinationSquares) > 0 {
		firstForward := pm.destinationSquares[0]
		if firstForward != nil && firstForward.Piece == nil {
			var further *Square
			furtherDirection := NeighborBelow
			if pm.piece.Side == sideWhite {
				furtherDirection = NeighborAbove
			}
			further = v.board.getNeighborSquare(firstForward, furtherDirection)

			if further != nil && further.Piece == nil {
				pm.destinationSquares = append(pm.destinationSquares, further)
			}
		}
	}

	rankForEnPassant := 5
	if pm.piece.Side == sideBlack {
		rankForEnPassant = 4
	}

	if pm.origin.Rank != rankForEnPassant {
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

			captureDirection := NeighborBelow
			if adj.Piece.Side == sideBlack {
				captureDirection = NeighborAbove
			}

			captureTarget := v.board.getNeighborSquare(adj, captureDirection)
			if captureTarget != nil {
				pm.destinationSquares = append(pm.destinationSquares, captureTarget)
			}
		}
	}
}
