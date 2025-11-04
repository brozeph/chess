package chess

import (
	"crypto/md5"
	"encoding/base64"
	"strconv"
	"strings"
)

// Game represents the state of a single chess game.
// It manages the board, move history, captured pieces, and turn sequence.
type Game struct {
	// Board is the current chessboard state.
	Board *Board
	// CaptureHistory is a list of pieces that have been captured.
	CaptureHistory []*Piece
	// MoveHistory is a chronological record of all moves made in the game.
	MoveHistory  []*moveEvent
	sideResolver func(*Game) Side
	emitter      eventEmitter
}

// createGame initializes a new Game with a standard starting board and hooks up board events.
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

// on registers an event handler for a given game event.
func (g *Game) on(event string, handler func(interface{})) {
	g.emitter.on(event, handler)
}

// emit triggers a game event with the given data.
func (g *Game) emit(event string, data interface{}) {
	g.emitter.emit(event, data)
}

// hookBoardEvents sets up listeners for events from the Board object.
// It bubbles up board-level events (like move, capture, etc.) to the game level.
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

		g.Board.LastMovedPiece = nil
		if len(g.MoveHistory) > 0 {
			g.Board.LastMovedPiece = g.MoveHistory[len(g.MoveHistory)-1].Piece
		}

		g.emit("undo", mv)
	})
}

// getCurrentSide determines which side (White or Black) has the current turn.
// It uses the sideResolver function if provided, otherwise defaults to alternating turns starting with White.
func (g *Game) getCurrentSide() Side {
	if g.sideResolver != nil {
		return g.sideResolver(g)
	}

	if len(g.MoveHistory)%2 == 0 {
		return sideWhite
	}
	return sideBlack
}

// getHashCode generates a unique hash for the current board position.
// This is used to detect position repetitions for threefold repetition draws.
func (g *Game) getHashCode() string {
	var builder strings.Builder

	for _, sq := range g.Board.Squares {
		if sq.Piece != nil {
			if builder.Len() > 0 {
				builder.WriteRune('-')
			}

			builder.WriteRune(sq.File)
			builder.WriteString(strconv.Itoa(sq.Rank))
			sideMarker := "b"
			if sq.Piece.Side == sideWhite {
				sideMarker = "w"
			}

			builder.WriteString(sideMarker)
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

// recordMove adds a move to the game's history and updates the capture history if a piece was taken.
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

// move is a wrapper around Board.Move that executes a move on the board.
func (g *Game) move(src, dest *Square, notation string) (*moveResult, error) {
	res, err := g.Board.Move(src, dest, false, notation)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// promote is a wrapper around Board.Promote that handles pawn promotion.
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
