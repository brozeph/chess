package chess

import "errors"

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

func (v *boardValidator) evaluateCastle(validMoves []potentialMoves) {
	getValidSquares := func(src *Square) []*Square {
		for _, vm := range validMoves {
			if vm.origin == src {
				return vm.destinationSquares
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
			canStepThroughD := !v.isSquareAttacked(squares['d'])
			res1.Undo()
			if canStepThroughD {
				res2, _ := v.board.Move(squares['e'], squares['c'], true, "")
				canLandOnC := !v.isSquareAttacked(squares['c'])
				res2.Undo()
				if canLandOnC {
					moveSquares := getValidSquares(squares['e'])
					if moveSquares != nil {
						moveSquares = append(moveSquares, squares['c'])
						for idx := range validMoves {
							if validMoves[idx].origin == squares['e'] {
								validMoves[idx].destinationSquares = moveSquares
							}
						}
					}
				}
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
			canStepThroughF := !v.isSquareAttacked(squares['f'])
			res1.Undo()
			if canStepThroughF {
				res2, _ := v.board.Move(squares['e'], squares['g'], true, "")
				canLandOnG := !v.isSquareAttacked(squares['g'])
				res2.Undo()
				if canLandOnG {
					moveSquares := getValidSquares(squares['e'])
					if moveSquares != nil {
						moveSquares = append(moveSquares, squares['g'])
						for idx := range validMoves {
							if validMoves[idx].origin == squares['e'] {
								validMoves[idx].destinationSquares = moveSquares
							}
						}
					}
				}
			}
		}
	}
}

func (v *boardValidator) filterKingAttack(kingSquare *Square, moves []potentialMoves) []potentialMoves {
	filtered := make([]potentialMoves, 0, len(moves))

	for _, mv := range moves {
		validSquares := []*Square{}
		for _, dest := range mv.destinationSquares {
			res, err := v.board.Move(mv.origin, dest, true, "")
			if err != nil {
				continue
			}

			isCheck := v.isSquareAttacked(kingSquare)
			if res.Move.Piece.Type == pieceKing {
				isCheck = v.isSquareAttacked(dest)
			}

			res.Undo()

			if !isCheck {
				validSquares = append(validSquares, dest)
			}
		}

		if len(validSquares) > 0 {
			filtered = append(filtered, potentialMoves{
				origin:             mv.origin,
				destinationSquares: validSquares,
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

func (v *boardValidator) Check() ([]potentialMoves, error) {
	if v.board == nil {
		return nil, errors.New("board is invalid")
	}

	squares := v.board.getSquares(v.game.getCurrentSide())
	validMoves := []potentialMoves{}
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
			validMoves = append(validMoves, potentialMoves{
				origin:             sq,
				destinationSquares: destSquares,
			})
		}
	}

	v.evaluateCastle(validMoves)
	validMoves = v.filterKingAttack(kingSquare, validMoves)

	attackers := v.findAttackers(kingSquare)
	for _, attacker := range attackers {
		data := &KingThreatEvent{
			AttackingSquare: attacker.square,
			KingSquare:      kingSquare,
		}

		if len(validMoves) == 0 {
			v.game.emit("checkmate", data)
			continue
		}
		v.game.emit("check", data)
	}

	return validMoves, nil
}
