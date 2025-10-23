package chess

type moveResult struct {
	Move *moveEvent
	undo func()
}

func (mr *moveResult) Undo() {
	if mr == nil || mr.undo == nil {
		return
	}

	mr.undo()
}

type potentialMoves struct {
	destinationSquares []*Square
	origin             *Square
	piece              *Piece
}
