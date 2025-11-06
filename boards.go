package chess

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Board represents the chess board and its state.
// It contains all the squares and the last moved piece.
// It can emit events for moves, captures, promotions, etc.
type Board struct {
	// Squares is a slice of 64 squares representing the board.
	Squares []*Square
	// LastMovedPiece points to the piece that was last moved.
	LastMovedPiece *Piece

	ev *eventHub
}

// createBoard initializes and returns a new Board with pieces in their standard
// starting positions.
func createBoard() *Board {
	b := &Board{
		Squares: make([]*Square, 0, 64),
		ev:      newEventHub(),
	}

	for i := 0; i < 64; i++ {
		f := "abcdefgh"[i%8]
		r := (i / 8) + 1
		sq := newSquare(rune(f), r)
		b.Squares = append(b.Squares, sq)

		switch r {
		case 1:
			switch i % 8 {
			case 0, 7:
				sq.Piece = newPiece(pieceRook, sideWhite)
			case 1, 6:
				sq.Piece = newPiece(pieceKnight, sideWhite)
			case 2, 5:
				sq.Piece = newPiece(pieceBishop, sideWhite)
			case 3:
				sq.Piece = newPiece(pieceQueen, sideWhite)
			default:
				sq.Piece = newPiece(pieceKing, sideWhite)
			}
		case 2:
			sq.Piece = newPiece(piecePawn, sideWhite)
		case 7:
			sq.Piece = newPiece(piecePawn, sideBlack)
		case 8:
			switch i % 8 {
			case 0, 7:
				sq.Piece = newPiece(pieceRook, sideBlack)
			case 1, 6:
				sq.Piece = newPiece(pieceKnight, sideBlack)
			case 2, 5:
				sq.Piece = newPiece(pieceBishop, sideBlack)
			case 3:
				sq.Piece = newPiece(pieceQueen, sideBlack)
			default:
				sq.Piece = newPiece(pieceKing, sideBlack)
			}
		}
	}

	return b
}

// loadBoard creates and returns a new Board from a Forsyth-Edwards Notation (FEN) string.
// It returns an error if the FEN string is invalid.
func loadBoard(fen string) (*Board, error) {
	prts := strings.Split(fen, " ")
	if len(prts) == 0 {
		return nil, fmt.Errorf("invalid FEN: %s", fen)
	}

	rs := strings.Split(prts[0], "/")
	if len(rs) != 8 {
		return nil, fmt.Errorf("invalid FEN ranks: %s", fen)
	}

	b := &Board{
		Squares: make([]*Square, 64),
		ev:      newEventHub(),
	}

	for ri, row := range rs {
		r := 8 - ri
		f := 0
		for _, ch := range row {
			if ch >= '1' && ch <= '8' {
				for n := 0; n < int(ch-'0'); n++ {
					idx := (r-1)*8 + f
					b.Squares[idx] = newSquare(rune("abcdefgh"[f]), r)
					f++
				}
				continue
			}

			if f >= 8 {
				return nil, fmt.Errorf("invalid FEN row: %s", row)
			}

			sq := newSquare(rune("abcdefgh"[f]), r)
			switch ch {
			case 'p':
				sq.Piece = newPiece(piecePawn, sideBlack)
			case 'P':
				sq.Piece = newPiece(piecePawn, sideWhite)
			case 'n':
				sq.Piece = newPiece(pieceKnight, sideBlack)
			case 'N':
				sq.Piece = newPiece(pieceKnight, sideWhite)
			case 'b':
				sq.Piece = newPiece(pieceBishop, sideBlack)
			case 'B':
				sq.Piece = newPiece(pieceBishop, sideWhite)
			case 'r':
				sq.Piece = newPiece(pieceRook, sideBlack)
			case 'R':
				sq.Piece = newPiece(pieceRook, sideWhite)
			case 'q':
				sq.Piece = newPiece(pieceQueen, sideBlack)
			case 'Q':
				sq.Piece = newPiece(pieceQueen, sideWhite)
			case 'k':
				sq.Piece = newPiece(pieceKing, sideBlack)
			case 'K':
				sq.Piece = newPiece(pieceKing, sideWhite)
			default:
				return nil, fmt.Errorf("invalid FEN piece: %c", ch)
			}

			idx := (r-1)*8 + f
			b.Squares[idx] = sq
			f++
		}
		if f != 8 {
			return nil, fmt.Errorf("invalid FEN row: %s", row)
		}
	}

	return b, nil
}

func (b *Board) emit(event string, data any) {
	if b == nil {
		return
	}

	b.ev.emit(event, data)
}

// fenPiecePlacement returns the piece placement data for Forsyth-Edwards Notation
func (b *Board) fenPiecePlacement() string {
	var fen strings.Builder
	ec := 0

	for r := 8; r >= 1; r-- {
		ec = 0
		for f := 'a'; f <= 'h'; f++ {
			sq := b.GetSquare(f, r)
			if sq.Piece == nil {
				ec++
				continue
			}

			if ec > 0 {
				fen.WriteString(strconv.Itoa(ec))
				ec = 0
			}

			fen.WriteString(sq.Piece.toFEN())
		}

		if ec > 0 {
			fen.WriteString(strconv.Itoa(ec))
		}

		if r > 1 {
			fen.WriteRune('/')
		}
	}

	return fen.String()
}

func (b *Board) getNeighborSquare(sq *Square, n neighbor) *Square {
	if sq == nil {
		return nil
	}

	switch sq.File {
	case 'a':
		if n == NeighborAboveLeft ||
			n == NeighborBelowLeft ||
			n == NeighborLeft {
			return nil
		}
	case 'b':
		if n == NeighborKnightLeftAbove ||
			n == NeighborKnightLeftBelow {
			return nil
		}
	case 'g':
		if n == NeighborKnightRightAbove ||
			n == NeighborKnightRightBelow {
			return nil
		}
	case 'h':
		if n == NeighborAboveRight ||
			n == NeighborBelowRight ||
			n == NeighborRight {
			return nil
		}
	}

	if sq.Rank == 1 && (n == NeighborBelow || n == NeighborBelowLeft || n == NeighborBelowRight ||
		n == NeighborKnightLeftBelow || n == NeighborKnightRightBelow || n == NeighborKnightBelowLeft || n == NeighborKnightBelowRight) {
		return nil
	}

	if sq.Rank == 8 && (n == NeighborAbove || n == NeighborAboveLeft || n == NeighborAboveRight ||
		n == NeighborKnightLeftAbove || n == NeighborKnightRightAbove || n == NeighborKnightAboveLeft || n == NeighborKnightAboveRight) {
		return nil
	}

	idx := b.indexOf(sq)
	if idx == -1 {
		return nil
	}

	trgt := idx + int(n)
	if trgt < 0 || trgt >= len(b.Squares) {
		return nil
	}

	return b.Squares[trgt]
}

func (b *Board) getSquareByName(nm string) *Square {
	if len(nm) != 2 {
		return nil
	}

	f := rune(nm[0])
	r := int(nm[1] - '0')

	return b.GetSquare(f, r)
}

func (b *Board) getSquares(sd Side) []*Square {
	res := make([]*Square, 0, 16)
	for _, sq := range b.Squares {
		if sq.Piece != nil && sq.Piece.Side == sd {
			res = append(res, sq)
		}
	}

	return res
}

func (b *Board) indexOf(sq *Square) int {
	if sq == nil {
		return -1
	}

	fi := int(sq.File - 'a')
	if fi < 0 || fi > 7 || sq.Rank < 1 || sq.Rank > 8 {
		return -1
	}

	return (sq.Rank-1)*8 + fi
}

// on registers an event handler for a given board event.
// The board supports the following events:
//   - "move":      emitted after a piece has been moved. The handler receives a *moveEvent.
//   - "capture":   emitted when a piece is captured. The handler receives the *moveEvent containing the captured piece.
//   - "castle":    emitted when a castling move is performed. The handler receives the *moveEvent.
//   - "enPassant": emitted when an en passant capture occurs. The handler receives the *moveEvent.
//   - "promote":   emitted when a pawn is promoted. The handler receives the promoted *Square.
//   - "undo":      emitted after a move has been undone. The handler receives the undone *moveEvent.
func (b *Board) on(e string, hndlr func(any)) {
	if b == nil {
		return
	}

	b.ev.on(e, hndlr)
}

// GetSquare returns the square at the given file and rank.
func (b *Board) GetSquare(f rune, r int) *Square {
	if r < 1 || r > 8 || f < 'a' || f > 'h' {
		return nil
	}

	idx := (r-1)*8 + int(f-'a')
	if idx < 0 || idx >= len(b.Squares) {
		return nil
	}

	return b.Squares[idx]
}

// Move performs a move on the board from a source square to a destination square.
// If simulate is true, the move is not committed to the board's history and no events are emitted.
// The returned moveResult contains an `undo` function that can be called to revert the move.
// It returns an error if the move is invalid.
func (b *Board) Move(src, dst *Square, sim bool, not ...string) (*moveResult, error) {
	if src == nil || dst == nil {
		return nil, errors.New("source and destination squares are required")
	}

	if src.Piece == nil {
		return nil, fmt.Errorf("no piece on source square %s", src.name())
	}

	n := ""
	if len(not) > 0 {
		n = not[0]
	}

	mv := &MoveEvent{
		Algebraic:     n,
		CapturedPiece: dst.Piece,
		PostSquare:    dst,
		PrevSquare:    src,
		Piece:         src.Piece,
		prevMoveCount: src.Piece.MoveCount,
		simulate:      sim,
	}

	dst.Piece = src.Piece
	src.Piece = nil

	mv.Castle = mv.Piece.Type == pieceKing && mv.prevMoveCount == 0 && (dst.File == 'g' || dst.File == 'c')
	mv.EnPassant = mv.Piece.Type == piecePawn && mv.CapturedPiece == nil && dst.File != mv.PrevSquare.File

	if mv.EnPassant {
		cs := b.GetSquare(dst.File, mv.PrevSquare.Rank)
		if cs != nil {
			mv.CapturedPiece = cs.Piece
			mv.EnPassantCaptureSquare = cs
			cs.Piece = nil
		}
	}

	if mv.Castle {
		rs := b.GetSquare('a', dst.Rank)
		rd := b.GetSquare('d', dst.Rank)
		if dst.File == 'g' {
			rs = b.GetSquare('h', dst.Rank)
			rd = b.GetSquare('f', dst.Rank)
		}

		if rs == nil || rs.Piece == nil {
			mv.Castle = false
		}

		if mv.Castle {
			mv.RookSource = rs
			mv.RookDestination = rd
			if rd != nil {
				rd.Piece = rs.Piece
				rs.Piece = nil
			}
		}
	}

	if !sim {
		mv.Piece.MoveCount++
		b.LastMovedPiece = mv.Piece
		b.emit("move", mv)
		if mv.CapturedPiece != nil {
			b.emit("capture", mv)
		}
		if mv.Castle {
			b.emit("castle", mv)
		}
		if mv.EnPassant {
			b.emit("enPassant", mv)
		}
	}

	undo := func() {
		if mv.undone {
			return
		}

		mv.PrevSquare.Piece = mv.Piece
		mv.PostSquare.Piece = mv.CapturedPiece

		if mv.EnPassant && mv.EnPassantCaptureSquare != nil {
			mv.EnPassantCaptureSquare.Piece = mv.CapturedPiece
			mv.PostSquare.Piece = nil
		}

		if mv.Castle && mv.RookSource != nil && mv.RookDestination != nil {
			mv.RookSource.Piece = mv.RookDestination.Piece
			mv.RookDestination.Piece = nil
		}

		if !mv.simulate {
			mv.Piece.MoveCount = mv.prevMoveCount
			b.LastMovedPiece = nil
			mv.undone = true
			b.emit("undo", mv)
			return
		}

		mv.undone = true
	}

	return &moveResult{
		Move: mv,
		undo: undo,
	}, nil
}

// Promote replaces the piece on a given square with a new piece.
// This is used for pawn promotion. It emits a "promote" event.
func (b *Board) Promote(sq *Square, p *Piece) (*Square, error) {
	if sq == nil {
		return nil, errors.New("square is required for promotion")
	}

	if sq.Piece == nil {
		return nil, fmt.Errorf("no piece to promote on %s", sq.name())
	}

	p.MoveCount = sq.Piece.MoveCount
	sq.Piece = p
	b.LastMovedPiece = p
	b.emit("promote", sq)

	return sq, nil
}
