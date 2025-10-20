package chess

type Move struct {
	Algebraic              string
	CapturedPiece          *Piece
	Castle                 bool
	EnPassant              bool
	Piece                  *Piece
	PostSquare             *Square
	PrevSquare             *Square
	Promotion              bool
	RookSource             *Square
	RookDestination        *Square
	EnPassantCaptureSquare *Square
	hashCode               string
	prevMoveCount          int
	simulate               bool
	undone                 bool
}

type MoveResult struct {
	Move *Move
	undo func()
}

func (mr *MoveResult) Undo() {
	if mr == nil || mr.undo == nil {
		return
	}

	mr.undo()
}
