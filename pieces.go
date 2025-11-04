package chess

import "strings"

// pieceType is an enumeration representing the type of a chess piece.
type pieceType int

const (
	pieceBishop pieceType = iota // A bishop.
	pieceKing                    // A king.
	pieceKnight                  // A knight.
	piecePawn                    // A pawn.
	pieceQueen                   // A queen.
	pieceRook                    // A rook.
)

// Piece represents a single chess piece on the board.
type Piece struct {
	// Type is the type of the piece (e.g., Pawn, Rook, King).
	Type pieceType
	// Side is the color of the piece (White or Black).
	Side Side
	// Notation is the standard algebraic notation for the piece (e.g., "R" for Rook).
	// Pawns have an empty string.
	Notation string
	// MoveCount tracks how many times the piece has moved. This is important for castling and pawn's first move.
	MoveCount int
}

// newPiece is a factory function that creates and returns a new Piece.
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

// toFEN converts the piece to its Forsyth-Edwards Notation (FEN) character.
// White pieces are uppercase (e.g., 'R'), and black pieces are lowercase (e.g., 'r').
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

// AlgebraicSymbol returns the rune used to represent the piece in various notations.
// White pieces are uppercase (e.g., 'P'), and black pieces are lowercase (e.g., 'p').
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
