package chess

import (
	"errors"
)

type validMove struct {
	Src     *Square
	Squares []*Square
}

type attackContext struct {
	attacked bool
	piece    *Piece
	square   *Square
}

type kingThreatEvent struct {
	AttackingSquare *Square
	KingSquare      *Square
}

type validationContext struct {
	destSquares []*Square
	origin      *Square
	p           *Piece
}

type validationResult struct {
	IsCheck      bool
	IsCheckmate  bool
	IsRepetition bool
	IsStalemate  bool
	ValidMoves   []validMove
}

type boardValidator struct {
	game  *Game
	board *Board
}

func CreateBoardValidator(g *Game) *boardValidator {
	return &boardValidator{
		game:  g,
		board: g.Board,
	}
}

func (v *boardValidator) evaluateCastle(validMoves []validMove) {
	getValidSquares := func(src *Square) []*Square {
		for _, vm := range validMoves {
			if vm.Src == src {
				return vm.Squares
			}
		}
		return nil
	}

	rank := 1
	if v.game.getCurrentSide() == sideBlack {
		rank = 8
	}

	squares := map[rune]*Square{
		'a': v.board.GetSquare('a', rank),
		'b': v.board.GetSquare('b', rank),
		'c': v.board.GetSquare('c', rank),
		'd': v.board.GetSquare('d', rank),
		'e': v.board.GetSquare('e', rank),
		'f': v.board.GetSquare('f', rank),
		'g': v.board.GetSquare('g', rank),
		'h': v.board.GetSquare('h', rank),
	}

	kingSquare := squares['e']
	if kingSquare == nil || kingSquare.Piece == nil {
		return
	}
	if kingSquare.Piece.Type != pieceKing || kingSquare.Piece.MoveCount != 0 {
		return
	}
	if v.isSquareAttacked(kingSquare) {
		return
	}

	// queen side
	if squares['a'] != nil &&
		squares['a'].Piece != nil &&
		squares['a'].Piece.Type == pieceRook &&
		squares['a'].Piece.MoveCount == 0 {

		if squares['b'].Piece == nil &&
			squares['c'].Piece == nil &&
			squares['d'].Piece == nil {

			res1, _ := v.board.Move(squares['e'], squares['d'], true, "")
			if !v.isSquareAttacked(squares['d']) {
				res1.Undo()
				res2, _ := v.board.Move(squares['e'], squares['c'], true, "")
				if !v.isSquareAttacked(squares['c']) {
					moveSquares := getValidSquares(squares['e'])
					if moveSquares != nil {
						moveSquares = append(moveSquares, squares['c'])
						for idx := range validMoves {
							if validMoves[idx].Src == squares['e'] {
								validMoves[idx].Squares = moveSquares
							}
						}
					}
				}
				res2.Undo()
			} else {
				res1.Undo()
			}
		}
	}

	// king side
	if squares['h'] != nil &&
		squares['h'].Piece != nil &&
		squares['h'].Piece.Type == pieceRook &&
		squares['h'].Piece.MoveCount == 0 {

		if squares['f'].Piece == nil &&
			squares['g'].Piece == nil {

			res1, _ := v.board.Move(squares['e'], squares['f'], true, "")
			if !v.isSquareAttacked(squares['f']) {
				res1.Undo()
				res2, _ := v.board.Move(squares['e'], squares['g'], true, "")
				if !v.isSquareAttacked(squares['g']) {
					moveSquares := getValidSquares(squares['e'])
					if moveSquares != nil {
						moveSquares = append(moveSquares, squares['g'])
						for idx := range validMoves {
							if validMoves[idx].Src == squares['e'] {
								validMoves[idx].Squares = moveSquares
							}
						}
					}
				}
				res2.Undo()
			} else {
				res1.Undo()
			}
		}
	}
}

func (v *boardValidator) filterKingAttack(kingSquare *Square, moves []validMove) []validMove {
	filtered := make([]validMove, 0, len(moves))

	for _, mv := range moves {
		validSquares := []*Square{}
		for _, dest := range mv.Squares {
			res, err := v.board.Move(mv.Src, dest, true, "")
			if err != nil {
				continue
			}

			isCheck := false
			if res.Move.Piece.Type != pieceKing {
				isCheck = v.isSquareAttacked(kingSquare)
			} else {
				isCheck = v.isSquareAttacked(dest)
			}

			res.Undo()

			if !isCheck {
				validSquares = append(validSquares, dest)
			}
		}

		if len(validSquares) > 0 {
			filtered = append(filtered, validMove{
				Src:     mv.Src,
				Squares: validSquares,
			})
		}
	}

	return filtered
}

func (v *boardValidator) findAttackers(target *Square) []attackContext {
	if target == nil || target.Piece == nil {
		return []attackContext{}
	}

	results := []attackContext{}

	checkDirection := func(nb neighbor) {
		current := v.board.getNeighborSquare(target, nb)
		for current != nil {
			if current.Piece != nil {
				if current.Piece.Side != target.Piece.Side {
					validator := CreatePieceValidator(current.Piece.Type, v.board)
					destSquares, err := validator.Check(current)
					if err == nil {
						for _, sq := range destSquares {
							if sq == target {
								results = append(results, attackContext{
									attacked: true,
									piece:    current.Piece,
									square:   current,
								})
								return
							}
						}
					}
				}
				return
			}
			current = v.board.getNeighborSquare(current, nb)
		}
	}

	checkKnight := func(nb neighbor) {
		current := v.board.getNeighborSquare(target, nb)
		if current == nil || current.Piece == nil {
			return
		}
		if current.Piece.Type != pieceKnight || current.Piece.Side == target.Piece.Side {
			return
		}
		validator := CreatePieceValidator(pieceKnight, v.board)
		destSquares, err := validator.Check(current)
		if err != nil {
			return
		}
		for _, sq := range destSquares {
			if sq == target {
				results = append(results, attackContext{
					attacked: true,
					piece:    current.Piece,
					square:   current,
				})
				return
			}
		}
	}

	directions := []neighbor{
		NeighborAbove,
		NeighborAboveRight,
		NeighborRight,
		NeighborBelowRight,
		NeighborBelow,
		NeighborBelowLeft,
		NeighborLeft,
		NeighborAboveLeft,
	}

	for _, dir := range directions {
		checkDirection(dir)
	}

	knightDirections := []neighbor{
		NeighborKnightAboveRight,
		NeighborKnightRightAbove,
		NeighborKnightBelowRight,
		NeighborKnightRightBelow,
		NeighborKnightBelowLeft,
		NeighborKnightLeftBelow,
		NeighborKnightAboveLeft,
		NeighborKnightLeftAbove,
	}

	for _, dir := range knightDirections {
		checkKnight(dir)
	}

	return results
}

func (v *boardValidator) isSquareAttacked(sq *Square) bool {
	return len(v.findAttackers(sq)) > 0
}

func (v *boardValidator) Check() ([]validMove, error) {
	if v.board == nil {
		return nil, errors.New("board is invalid")
	}

	squares := v.board.getSquares(v.game.getCurrentSide())
	validMoves := []validMove{}
	var kingSquare *Square

	for _, sq := range squares {
		if sq.Piece == nil {
			continue
		}

		if sq.Piece.Type == pieceKing {
			kingSquare = sq
		}

		validator := CreatePieceValidator(sq.Piece.Type, v.board)
		destSquares, err := validator.Check(sq)
		if err != nil {
			return nil, err
		}

		if len(destSquares) > 0 {
			validMoves = append(validMoves, validMove{
				Src:     sq,
				Squares: destSquares,
			})
		}
	}

	v.evaluateCastle(validMoves)
	validMoves = v.filterKingAttack(kingSquare, validMoves)

	attackers := v.findAttackers(kingSquare)
	for _, attacker := range attackers {
		data := kingThreatEvent{
			AttackingSquare: attacker.square,
			KingSquare:      kingSquare,
		}
		if len(validMoves) == 0 {
			v.game.emit("checkmate", data)
		} else {
			v.game.emit("check", data)
		}
	}

	return validMoves, nil
}

type gameValidator struct {
	game *Game
}

func CreateGameValidator(g *Game) *gameValidator {
	return &gameValidator{game: g}
}

func (gv *gameValidator) findKingSquare(sd Side) *Square {
	squares := gv.game.Board.getSquares(sd)
	for _, sq := range squares {
		if sq.Piece != nil && sq.Piece.Type == pieceKing {
			return sq
		}
	}
	return nil
}

func (gv *gameValidator) isRepetition() bool {
	counts := map[string]int{}
	for _, mv := range gv.game.MoveHistory {
		counts[mv.hashCode]++
		if counts[mv.hashCode] == 3 {
			return true
		}
	}
	return false
}

func (gv *gameValidator) Check() (*validationResult, error) {
	result := &validationResult{
		IsCheck:      false,
		IsCheckmate:  false,
		IsRepetition: false,
		IsStalemate:  false,
		ValidMoves:   []validMove{},
	}

	bv := CreateBoardValidator(gv.game)
	validMoves, err := bv.Check()
	if err != nil {
		return nil, err
	}

	kingSquare := gv.findKingSquare(gv.game.getCurrentSide())
	isAttacked := bv.isSquareAttacked(kingSquare)

	result.IsCheck = isAttacked && len(validMoves) > 0
	result.IsCheckmate = isAttacked && len(validMoves) == 0
	result.IsStalemate = !isAttacked && len(validMoves) == 0
	result.ValidMoves = validMoves
	result.IsRepetition = gv.isRepetition()

	return result, nil
}

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
