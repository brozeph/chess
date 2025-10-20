package chess

import "fmt"

type neighbor int

const (
	neighborAbove            neighbor = 8
	neighborAboveLeft        neighbor = 7
	neighborAboveRight       neighbor = 9
	neighborBelow            neighbor = -8
	neighborBelowLeft        neighbor = -9
	neighborBelowRight       neighbor = -7
	neighborLeft             neighbor = -1
	neighborRight            neighbor = 1
	neighborKnightAboveLeft  neighbor = 15
	neighborKnightAboveRight neighbor = 17
	neighborKnightBelowLeft  neighbor = -17
	neighborKnightBelowRight neighbor = -15
	neighborKnightLeftAbove  neighbor = 6
	neighborKnightLeftBelow  neighbor = -10
	neighborKnightRightAbove neighbor = 10
	neighborKnightRightBelow neighbor = -6
)

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
