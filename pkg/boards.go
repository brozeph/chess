package chess

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Board struct {
	Squares        []*Square
	LastMovedPiece *Piece
	emitter        eventEmitter
}

func CreateBoard() *Board {
	b := &Board{
		Squares: make([]*Square, 0, 64),
		emitter: newEventEmitter(),
	}

	for i := 0; i < 64; i++ {
		file := "abcdefgh"[i%8]
		rank := (i / 8) + 1
		sq := newSquare(rune(file), rank)
		b.Squares = append(b.Squares, sq)

		switch rank {
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

func LoadBoard(fen string) (*Board, error) {
	parts := strings.Split(fen, " ")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid FEN: %s", fen)
	}

	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("invalid FEN ranks: %s", fen)
	}

	b := &Board{
		Squares: make([]*Square, 64),
		emitter: newEventEmitter(),
	}
	for rankIdx, row := range ranks {
		rank := 8 - rankIdx
		file := 0
		for _, ch := range row {
			if ch >= '1' && ch <= '8' {
				for n := 0; n < int(ch-'0'); n++ {
					idx := (rank-1)*8 + file
					b.Squares[idx] = newSquare(rune("abcdefgh"[file]), rank)
					file++
				}
				continue
			}

			if file >= 8 {
				return nil, fmt.Errorf("invalid FEN row: %s", row)
			}

			sq := newSquare(rune("abcdefgh"[file]), rank)
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

			idx := (rank-1)*8 + file
			b.Squares[idx] = sq
			file++
		}
		if file != 8 {
			return nil, fmt.Errorf("invalid FEN row: %s", row)
		}
	}

	return b, nil
}

func (b *Board) on(event string, handler func(interface{})) {
	b.emitter.on(event, handler)
}

func (b *Board) emit(event string, data interface{}) {
	b.emitter.emit(event, data)
}

func (b *Board) indexOf(sq *Square) int {
	if sq == nil {
		return -1
	}
	fileIdx := int(sq.File - 'a')
	if fileIdx < 0 || fileIdx > 7 || sq.Rank < 1 || sq.Rank > 8 {
		return -1
	}
	return (sq.Rank-1)*8 + fileIdx
}

func (b *Board) GetSquare(file rune, rank int) *Square {
	if rank < 1 || rank > 8 || file < 'a' || file > 'h' {
		return nil
	}
	idx := (rank-1)*8 + int(file-'a')
	if idx < 0 || idx >= len(b.Squares) {
		return nil
	}
	return b.Squares[idx]
}

func (b *Board) getSquareByName(name string) *Square {
	if len(name) != 2 {
		return nil
	}
	file := rune(name[0])
	rank := int(name[1] - '0')
	return b.GetSquare(file, rank)
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

	target := idx + int(n)
	if target < 0 || target >= len(b.Squares) {
		return nil
	}

	return b.Squares[target]
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

func (b *Board) FEN() string {
	var fen strings.Builder
	emptyCount := 0

	for rank := 8; rank >= 1; rank-- {
		emptyCount = 0
		for file := 'a'; file <= 'h'; file++ {
			sq := b.GetSquare(file, rank)
			if sq.Piece == nil {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				fen.WriteString(sq.Piece.toFEN())
			}
		}
		if emptyCount > 0 {
			fen.WriteString(strconv.Itoa(emptyCount))
		}
		if rank > 1 {
			fen.WriteRune('/')
		}
	}

	return fen.String()
}

func (b *Board) Move(src, dest *Square, simulate bool, notation string) (*MoveResult, error) {
	if src == nil || dest == nil {
		return nil, errors.New("source and destination squares are required")
	}

	if src.Piece == nil {
		return nil, fmt.Errorf("no piece on source square %s", src.name())
	}

	mv := &Move{
		Algebraic:     notation,
		CapturedPiece: dest.Piece,
		PostSquare:    dest,
		PrevSquare:    src,
		Piece:         src.Piece,
		prevMoveCount: src.Piece.MoveCount,
		simulate:      simulate,
	}

	dest.Piece = src.Piece
	src.Piece = nil

	mv.Castle = mv.Piece.Type == pieceKing && mv.prevMoveCount == 0 && (dest.File == 'g' || dest.File == 'c')
	mv.EnPassant = mv.Piece.Type == piecePawn && mv.CapturedPiece == nil && dest.File != mv.PrevSquare.File

	if mv.EnPassant {
		captureSq := b.GetSquare(dest.File, mv.PrevSquare.Rank)
		if captureSq != nil {
			mv.CapturedPiece = captureSq.Piece
			mv.EnPassantCaptureSquare = captureSq
			captureSq.Piece = nil
		}
	}

	if mv.Castle {
		var rookSource, rookDest *Square
		if dest.File == 'g' {
			rookSource = b.GetSquare('h', dest.Rank)
			rookDest = b.GetSquare('f', dest.Rank)
		} else {
			rookSource = b.GetSquare('a', dest.Rank)
			rookDest = b.GetSquare('d', dest.Rank)
		}

		if rookSource == nil || rookSource.Piece == nil {
			mv.Castle = false
		} else {
			mv.RookSource = rookSource
			mv.RookDestination = rookDest
			if rookDest != nil {
				rookDest.Piece = rookSource.Piece
				rookSource.Piece = nil
			}
		}
	}

	if !simulate {
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

	return &MoveResult{
		Move: mv,
		undo: undo,
	}, nil
}

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
