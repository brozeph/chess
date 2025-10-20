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
	specialValidator func(v *pieceValidator, ctx *validationContext)
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
		v.specialValidator = knightSpecial
		v.repeat = 1
	case piecePawn:
		v.allowForward = true
		v.specialValidator = pawnSpecial
		v.repeat = 1
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

	ctx := &validationContext{
		destSquares: []*Square{},
		origin:      origin,
		p:           origin.Piece,
	}

	findMoveOptions := func(direction neighbor) {
		current := v.board.getNeighborSquare(ctx.origin, direction)
		steps := 0

		for current != nil && steps < v.repeat {
			block := current.Piece != nil && (ctx.p.Type == piecePawn || current.Piece.Side == ctx.p.Side)
			capture := current.Piece != nil && !block

			if !block {
				ctx.destSquares = append(ctx.destSquares, current)
			}

			if capture || block {
				break
			}

			current = v.board.getNeighborSquare(current, direction)
			steps++
		}
	}

	if v.allowForward {
		if ctx.p.Side == sideWhite {
			findMoveOptions(NeighborAbove)
		} else {
			findMoveOptions(NeighborBelow)
		}
	}

	if v.allowBackward {
		if ctx.p.Side == sideWhite {
			findMoveOptions(NeighborBelow)
		} else {
			findMoveOptions(NeighborAbove)
		}
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

	return ctx.destSquares, nil
}

func knightSpecial(v *pieceValidator, ctx *validationContext) {
	candidates := []*Square{}

	aboveLeft := v.board.getNeighborSquare(ctx.origin, NeighborAboveLeft)
	aboveRight := v.board.getNeighborSquare(ctx.origin, NeighborAboveRight)
	belowLeft := v.board.getNeighborSquare(ctx.origin, NeighborBelowLeft)
	belowRight := v.board.getNeighborSquare(ctx.origin, NeighborBelowRight)

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

		if sq.Piece == nil || sq.Piece.Side != ctx.p.Side {
			ctx.destSquares = append(ctx.destSquares, sq)
		}
	}
}

func pawnSpecial(v *pieceValidator, ctx *validationContext) {
	var diagLeft, diagRight *Square
	if ctx.p.Side == sideWhite {
		diagLeft = v.board.getNeighborSquare(ctx.origin, NeighborAboveLeft)
		diagRight = v.board.getNeighborSquare(ctx.origin, NeighborAboveRight)
	} else {
		diagLeft = v.board.getNeighborSquare(ctx.origin, NeighborBelowLeft)
		diagRight = v.board.getNeighborSquare(ctx.origin, NeighborBelowRight)
	}

	for _, sq := range []*Square{diagLeft, diagRight} {
		if sq == nil {
			continue
		}
		if sq.Piece != nil && sq.Piece.Side != ctx.p.Side {
			ctx.destSquares = append(ctx.destSquares, sq)
		}
	}

	if ctx.p.MoveCount == 0 && len(ctx.destSquares) > 0 {
		firstForward := ctx.destSquares[0]
		if firstForward != nil && firstForward.Piece == nil {
			var further *Square
			if ctx.p.Side == sideWhite {
				further = v.board.getNeighborSquare(firstForward, NeighborAbove)
			} else {
				further = v.board.getNeighborSquare(firstForward, NeighborBelow)
			}

			if further != nil && further.Piece == nil {
				ctx.destSquares = append(ctx.destSquares, further)
			}
		}
	}

	rankForEnPassant := 5
	if ctx.p.Side == sideBlack {
		rankForEnPassant = 4
	}

	if ctx.origin.Rank != rankForEnPassant {
		return
	}

	left := v.board.getNeighborSquare(ctx.origin, NeighborLeft)
	right := v.board.getNeighborSquare(ctx.origin, NeighborRight)

	for _, adj := range []*Square{left, right} {
		if adj == nil || adj.Piece == nil {
			continue
		}

		if adj.Piece.Type == piecePawn &&
			adj.Piece.Side != ctx.p.Side &&
			adj.Piece.MoveCount == 1 &&
			v.board.LastMovedPiece == adj.Piece {
			var captureTarget *Square
			if adj.Piece.Side == sideBlack {
				captureTarget = v.board.getNeighborSquare(adj, NeighborAbove)
			} else {
				captureTarget = v.board.getNeighborSquare(adj, NeighborBelow)
			}
			if captureTarget != nil {
				ctx.destSquares = append(ctx.destSquares, captureTarget)
			}
		}
	}
}
