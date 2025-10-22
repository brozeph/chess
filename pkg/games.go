package chess

import (
	"crypto/md5"
	"encoding/base64"
	"strconv"
	"strings"
)

type Game struct {
	Board          *Board
	CaptureHistory []*Piece
	MoveHistory    []*moveEvent
	sideResolver   func(*Game) Side
	emitter        eventEmitter
}

func createGame() *Game {
	g := &Game{
		Board:          CreateBoard(),
		CaptureHistory: []*Piece{},
		MoveHistory:    []*moveEvent{},
		emitter:        newEventEmitter(),
	}

	g.hookBoardEvents()

	return g
}

func (g *Game) on(event string, handler func(interface{})) {
	g.emitter.on(event, handler)
}

func (g *Game) emit(event string, data interface{}) {
	g.emitter.emit(event, data)
}

func (g *Game) hookBoardEvents() {
	g.Board.on("move", func(data interface{}) {
		mv, ok := data.(*moveEvent)
		if !ok || mv == nil {
			return
		}
		g.recordMove(mv)
		g.emit("move", mv)
	})

	g.Board.on("capture", func(data interface{}) {
		g.emit("capture", data)
	})

	g.Board.on("castle", func(data interface{}) {
		g.emit("castle", data)
	})

	g.Board.on("enPassant", func(data interface{}) {
		g.emit("enPassant", data)
	})

	g.Board.on("promote", func(data interface{}) {
		g.emit("promote", data)
	})

	g.Board.on("undo", func(data interface{}) {
		mv, ok := data.(*moveEvent)
		if !ok || mv == nil {
			return
		}

		if len(g.MoveHistory) > 0 {
			g.MoveHistory = g.MoveHistory[:len(g.MoveHistory)-1]
		}

		if mv.CapturedPiece != nil && len(g.CaptureHistory) > 0 {
			g.CaptureHistory = g.CaptureHistory[:len(g.CaptureHistory)-1]
		}

		if len(g.MoveHistory) > 0 {
			g.Board.LastMovedPiece = g.MoveHistory[len(g.MoveHistory)-1].Piece
		} else {
			g.Board.LastMovedPiece = nil
		}

		g.emit("undo", mv)
	})
}

func (g *Game) getCurrentSide() Side {
	if g.sideResolver != nil {
		return g.sideResolver(g)
	}

	if len(g.MoveHistory)%2 == 0 {
		return sideWhite
	}
	return sideBlack
}

func (g *Game) getHashCode() string {
	var builder strings.Builder

	for _, sq := range g.Board.Squares {
		if sq.Piece != nil {
			if builder.Len() > 0 {
				builder.WriteRune('-')
			}
			builder.WriteRune(sq.File)
			builder.WriteString(strconv.Itoa(sq.Rank))
			if sq.Piece.Side == sideWhite {
				builder.WriteString("w")
			} else {
				builder.WriteString("b")
			}
			builder.WriteString(sq.Piece.Notation)
			if sq.Piece.Type == piecePawn {
				builder.WriteString("p")
			}
		}
	}

	sum := builder.String()
	hash := md5.Sum([]byte(sum))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (g *Game) recordMove(mv *moveEvent) {
	if mv == nil {
		return
	}

	mv.hashCode = g.getHashCode()

	g.MoveHistory = append(g.MoveHistory, mv)
	if mv.CapturedPiece != nil {
		g.CaptureHistory = append(g.CaptureHistory, mv.CapturedPiece)
	}
}

func (g *Game) move(src, dest *Square, notation string) (*moveResult, error) {
	res, err := g.Board.Move(src, dest, false, notation)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (g *Game) promote(sq *Square, p *Piece) (*Square, error) {
	target, err := g.Board.Promote(sq, p)
	if err != nil {
		return nil, err
	}

	if len(g.MoveHistory) > 0 {
		g.MoveHistory[len(g.MoveHistory)-1].Promotion = true
	}

	return target, nil
}
