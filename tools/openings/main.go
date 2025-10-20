package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	openingsCSV         = "./data/openings.csv"
	startingPositionFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
)

var (
	// group 1: piece (and optional from square (e.g., Nf, e, R1, etc.))
	// group 2: capture indicator (x)
	// group 3: to file (a-h)
	// group 4: to rank (1-8)
	reAlgebraicMove = regexp.MustCompile(`([BKNPRQ]?[a-h1-8]?)(x?)([a-h])([1-8])`)
)

type opening struct {
	FEN      string
	Name     string
	Sequence string
}

func applyMove(fen, mv string, isWht bool) string {
	// FEN notation begins with the upper left corner of the board (a8)
	//  - each rank (row) in FEN notation is separated by a slash (/)
	//  - each piece is represented by a letter (uppercase for White, lowercase for Black)
	//  - empty squares are represented by numbers (1-8, where 1 means 1 empty square, 2 means 2 empty squares, etc.)
	//  - the order of pieces in FEN is: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
	//  - if the first file in the rank is empty, it is represented by a number at the start of the rank

	var fromF, fromR, toF, toR int
	var isCpt, isCstl bool
	var pc string

	// determine if move is a castle
	for _, tst := range []string{"O-O", "0-0", "O-O-O", "0-0-0"} {
		if strings.HasPrefix(mv, tst) {
			pc = "K"
			isCstl = true

			// check for King side castle
			if tst == "O-O" || tst == "0-0" {
				fromF = 4
				toF = 7
				break
			}

			// check for Queen side castle
			fromF = 4
			toF = 2
			break
		}
	}

	if isCstl {
		fmt.Printf("(%s) - Castling move for %s side\n", mv, map[bool]string{true: "White", false: "Black"}[isWht])
		return fen
	}

	// parse the algebraic move
	mtchs := reAlgebraicMove.FindStringSubmatch(mv)

	// determine which piece is moving and to where
	for i, prt := range mtchs {
		switch i {
		case 1:
			// piece and optional from square
			if prt == "" {
				pc = "P"
				continue
			}

			// pawn move with from file specified
			if len(prt) == 1 && prt[0] >= 'a' && prt[0] <= 'h' {
				pc = "P"
				fromF = int(prt[0]-'a') + 1
				continue
			}

			// pawn move with from rank specified
			if len(prt) == 1 && prt[0] >= '1' && prt[0] <= '8' {
				pc = "P"
				fromR = int(prt[0] - '0')
				continue
			}

			pc = string(prt[0])

			// check for optional from file or rank
			if len(prt) == 2 {
				if prt[1] >= 'a' && prt[1] <= 'h' {
					fromF = int(prt[1]-'a') + 1
					continue
				}

				fromR = int(prt[1] - '0')
			}
		case 2:
			// capture indicator
			isCpt = (prt == "x")
		case 3:
			// to file
			toF = int(prt[0]-'a') + 1
		case 4:
			// to rank
			toR = int(prt[0] - '0')
		}
	}

	// determine where the piece is moving from

	fmt.Printf(
		"(%s) - Moving piece %s from file %d, rank %d to file %d, rank %d (white: %t, capture: %t, castling: %t)\n",
		mv,
		pc,
		fromF,
		fromR,
		toF,
		toR,
		isWht,
		isCpt,
		isCstl)

	/*for i, rank := range rnks {
		// Process each rank (row) to apply the move
		// This is a simplified example and does not account for all chess rules
		fmt.Printf("Processing rank %d: %s\n", i, rank)
	}//*/

	return fen
}

func convertToFEN(seq string) string {
	fen := startingPositionFEN

	// break the sequence into moves (format is 1 e4 e5 2 Nf3 Nc6 ...)
	prts := strings.Fields(seq)

	for i, prt := range prts {
		// skip move numbers (the first in a sequence of 3 parts)
		if i%3 == 0 {
			fmt.Print("processing move number:", prt, "\n")
			continue
		}

		// process moves here to build FEN (not implemented)
		fen = applyMove(fen, prt, (i%2 != 0))
	}

	return fen
}

func readOpenings(pth string) ([]opening, error) {
	// read in openings from CSV file
	f, err := os.Open(pth)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var openings []opening
	for _, record := range records {
		// handle both 2 and 3 column formats
		if len(record) == 3 {
			openings = append(openings, opening{
				FEN:      record[0],
				Name:     record[2],
				Sequence: record[1],
			})

			continue
		}

		if len(record) == 2 {
			openings = append(openings, opening{
				Name:     record[1],
				Sequence: record[0],
			})

			continue
		}
	}

	return openings, nil
}

func main() {
	// read in openings from CSV file
	ops, err := readOpenings(openingsCSV)
	if err != nil {
		fmt.Printf("Error reading openings: %v\n", err)
		return
	}

	// process openings as needed
	for _, op := range ops {
		// if FEN is missing, convert sequence to FEN
		if op.FEN == "" {
			op.FEN = convertToFEN(op.Sequence)
		}

		fmt.Printf("Opening: %s, Sequence: %s, FEN: %s\n", op.Name, op.Sequence, op.FEN)
	}
}
