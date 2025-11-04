package chess

import "strings"

type pieceType int

const (
	pieceBishop pieceType = iota
	pieceKing
	pieceKnight
	piecePawn
	pieceQueen
	pieceRook
)

type Piece struct {
	Type      pieceType
	Side      Side
	Notation  string
	MoveCount int
}

func newPiece(pt pieceType, sd Side) *Piece {
	switch pt {
	case pieceBishop:
		return &Piece{Type: pieceBishop, Side: sd, Notation: "B"}
	case pieceKing:
		return &Piece{Type: pieceKing, Side: sd, Notation: "K"}
	case pieceKnight:
		return &Piece{Type: pieceKnight, Side: sd, Notation: "N"}
	case piecePawn:
		return &Piece{Type: piecePawn, Side: sd, Notation: ""}
	case pieceQueen:
		return &Piece{Type: pieceQueen, Side: sd, Notation: "Q"}
	case pieceRook:
		return &Piece{Type: pieceRook, Side: sd, Notation: "R"}
	default:
		return nil
	}
}

func (p *Piece) toFEN() string {
	if p == nil {
		return ""
	}

	symbol := ""
	switch p.Type {
	case piecePawn:
		symbol = "p"
	case pieceKnight:
		symbol = "n"
	case pieceBishop:
		symbol = "b"
	case pieceRook:
		symbol = "r"
	case pieceQueen:
		symbol = "q"
	case pieceKing:
		symbol = "k"
	}

	if p.Side == sideWhite {
		return strings.ToUpper(symbol)
	}

	return symbol
}

func (p *Piece) AlgebraicSymbol() rune {
	if p == nil {
		return '.'
	}

	var symbol rune
	switch p.Type {
	case piecePawn:
		symbol = 'p'
	case pieceKnight:
		symbol = 'n'
	case pieceBishop:
		symbol = 'b'
	case pieceRook:
		symbol = 'r'
	case pieceQueen:
		symbol = 'q'
	case pieceKing:
		symbol = 'k'
	default:
		symbol = '?'
	}

	if p.Side == sideWhite {
		return rune(strings.ToUpper(string(symbol))[0])
	}

	return symbol
}
