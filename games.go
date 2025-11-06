package chess

import (
	"crypto/md5"
	"encoding/base64"
	"math"
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
	MoveHistory []*MoveEvent

	cstl string
	enP  *Square
	ev   *eventHub
	hmc  int
	fmn  int
	wf   bool
}

// createGame initializes a new Game with a standard starting board and hooks up board events.
func createGame(wf ...bool) *Game {
	g := &Game{
		Board:          createBoard(),
		CaptureHistory: []*Piece{},
		MoveHistory:    []*MoveEvent{},
		cstl:           "KQkq", // default is both Kings can castle King and Queen side
		ev:             newEventHub(),
		fmn:            1,    // default is first move
		wf:             true, // default is white moves first
	}

	// check if game was created with a white first override
	if len(wf) > 0 {
		g.wf = wf[0]
	}

	g.hookBoardEvents()

	return g
}

// emit triggers a game event with the given data.
func (g *Game) emit(event string, data interface{}) {
	if g == nil {
		return
	}

	g.ev.emit(event, data)
}

func (g *Game) fen() string {
	var b strings.Builder

	// 1. Piece placement
	b.WriteString(g.Board.fenPiecePlacement())

	// add a space
	b.WriteRune(' ')

	// 2. Active color
	if g.getCurrentSide() == sideWhite {
		b.WriteRune('w')
	} else {
		b.WriteRune('b')
	}

	// add a space
	b.WriteRune(' ')

	// 3. Castling availability
	b.WriteString(g.cstl)

	// add a space
	b.WriteRune(' ')

	// 4. En passant target
	if g.enP != nil {
		b.WriteString(g.enP.name())
	} else {
		b.WriteRune('-')
	}

	// add a space
	b.WriteRune(' ')

	// 5. Halfmove clock
	b.WriteString(strconv.Itoa(g.hmc))

	// add a space
	b.WriteRune(' ')

	// 6. Fullmove number
	b.WriteString(strconv.Itoa(g.fmn))

	return b.String()
}

// getCurrentSide determines which side (White or Black) has the current turn.
func (g *Game) getCurrentSide() Side {
	if len(g.MoveHistory)%2 == 0 {
		if g.wf {
			return sideWhite
		}

		return sideBlack
	}

	if g.wf {
		return sideBlack
	}

	return sideWhite
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

// hookBoardEvents sets up listeners for events from the Board object.
// It bubbles up board-level events (like move, capture, etc.) to the game level.
func (g *Game) hookBoardEvents() {
	g.Board.on("move", func(data interface{}) {
		mv, ok := data.(*MoveEvent)
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
		mv, ok := data.(*MoveEvent)
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

// move is a wrapper around Board.Move that executes a move on the board.
func (g *Game) move(src, dest *Square, notation string) (*moveResult, error) {
	res, err := g.Board.Move(src, dest, false, notation)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// on registers an event handler for a given game event.
func (g *Game) on(e string, hndlr func(any)) {
	if g == nil {
		return
	}

	g.ev.on(e, hndlr)
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

// recordMove adds a move to the game's history and updates the capture history if a piece was taken.
func (g *Game) recordMove(mv *MoveEvent) {
	if mv == nil {
		return
	}

	// create the move history entry
	mv.hashCode = g.getHashCode()
	g.MoveHistory = append(g.MoveHistory, mv)
	if mv.CapturedPiece != nil {
		g.CaptureHistory = append(g.CaptureHistory, mv.CapturedPiece)
	}

	// update hmc (halfmove count) if appropriate
	if mv.CapturedPiece != nil || mv.Piece.Type == piecePawn {
		g.hmc = 0
	} else {
		g.hmc++
	}

	// update fmn (fullmove number) if appropriate
	if g.fmn >= 1 && g.getCurrentSide() == sideWhite {
		g.fmn++
	}

	// update cstl (castle availability) if appropriate
	// if the King has moved, casteling King and Queen side is disabled
	if mv.Piece.Type == pieceKing {
		if mv.Piece.Side == sideWhite {
			g.cstl = strings.ReplaceAll(g.cstl, "K", "")
			g.cstl = strings.ReplaceAll(g.cstl, "Q", "")
		}

		if mv.Piece.Side == sideBlack {
			g.cstl = strings.ReplaceAll(g.cstl, "k", "")
			g.cstl = strings.ReplaceAll(g.cstl, "q", "")
		}
	}

	// if a Rook has moved, check the previous square position
	// to determine which (King or Queen) castle option to remove
	if mv.Piece.Type == pieceRook {
		if mv.PrevSquare.File == 'a' {
			if mv.PrevSquare.Rank == 1 {
				g.cstl = strings.ReplaceAll(g.cstl, "Q", "")
			}

			if mv.PrevSquare.Rank == 8 {
				g.cstl = strings.ReplaceAll(g.cstl, "q", "")
			}
		}

		if mv.PrevSquare.File == 'h' {
			if mv.PrevSquare.Rank == 1 {
				g.cstl = strings.ReplaceAll(g.cstl, "K", "")
			}

			if mv.PrevSquare.Rank == 8 {
				g.cstl = strings.ReplaceAll(g.cstl, "k", "")
			}
		}
	}

	// unassign enP (enpassant target), and reset if appropriate
	g.enP = nil
	if mv.Piece.Type == piecePawn {
		// check to see if pawn moved 2 squares
		if mv.Piece.MoveCount == 1 {
			prvRnk := mv.PrevSquare.Rank
			dstRnk := mv.PostSquare.Rank

			if math.Abs(float64(dstRnk-prvRnk)) == 2 {
				// get the square over which the Pawn passed
				if dstRnk > prvRnk {
					g.enP = g.Board.GetSquare(mv.PrevSquare.File, dstRnk-1)
				}

				if dstRnk < prvRnk {
					g.enP = g.Board.GetSquare(mv.PrevSquare.File, dstRnk+1)
				}
			}
		}
	}
}
