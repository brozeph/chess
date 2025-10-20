package chess

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
