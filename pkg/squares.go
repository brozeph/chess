package chess

import "fmt"

type Square struct {
	File  rune
	Rank  int
	Piece *Piece
}

func newSquare(file rune, rank int) *Square {
	return &Square{File: file, Rank: rank}
}

func (sq *Square) name() string {
	return fmt.Sprintf("%c%d", sq.File, sq.Rank)
}
