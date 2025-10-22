package chess

type gameValidator struct {
	game *Game
}

type attackContext struct {
	attacked bool
	piece    *Piece
	square   *Square
}

type validationResult struct {
	IsCheck      bool
	IsCheckmate  bool
	IsRepetition bool
	IsStalemate  bool
	ValidMoves   []potentialMoves
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
		ValidMoves:   []potentialMoves{},
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
