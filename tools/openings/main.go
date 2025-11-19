package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/brozeph/chess"
)

const (
	openingsCSV = "./data/openings.csv"
)

type opening struct {
	ECO          string
	Moves        string
	Name         string
	ResultFEN    string
	SequenceFENs []string
}

func convertToFEN(mvs string) []string {
	c := chess.CreateAlgebraicGameClient()

	// break the sequence into moves (format is 1 e4 e5 2 Nf3 Nc6 ...)
	trns := strings.Fields(mvs)
	fens := []string{}

	for i, trn := range trns {
		// skip move numbers (the first char in a sequence of 3 parts "# white black")
		if i%3 == 0 {
			fmt.Print("processing turn number:", trn, "\n")
			continue
		}

		// process moves here to build FEN (not implemented)
		c.Move(trn)
		fens = append(fens, c.FEN())
	}

	return fens
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
	for i, record := range records {
		// skip header row
		if i == 0 {
			continue
		}

		// handle multiple column formats (legacy support)
		switch len(record) {
		case 5:
			openings = append(openings, opening{
				Moves:        record[0],
				ECO:          record[1],
				Name:         record[2],
				ResultFEN:    record[3],
				SequenceFENs: strings.Split(record[4], ","),
			})

		case 4:
			openings = append(openings, opening{
				Moves:        record[0],
				Name:         record[1],
				ResultFEN:    record[2],
				SequenceFENs: strings.Split(record[3], ","),
			})
		case 2:
			openings = append(openings, opening{
				Moves: record[0],
				Name:  record[1],
			})
		}
	}

	return openings, nil
}

func writeOpenings(ops []opening) error {
	// write updated openings to CSV file
	f, err := os.Create(openingsCSV)
	if err != nil {
		return fmt.Errorf("error creating openings file: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// write header row
	if err := writer.Write([]string{"Moves", "ECO", "Name", "ResultFEN", "SequenceFENs"}); err != nil {
		return fmt.Errorf("error writing header row: %w", err)
	}

	// write each opening
	for _, op := range ops {
		err := writer.Write([]string{op.Moves, op.ECO, op.Name, op.ResultFEN, strings.Join(op.SequenceFENs, ",")})
		if err != nil {
			return fmt.Errorf("error writing openings file: %w", err)
		}
	}

	return nil
}

func main() {
	// read in openings from CSV file
	ops, err := readOpenings(openingsCSV)
	if err != nil {
		fmt.Printf("Error reading openings: %v\n", err)
		return
	}

	// process openings as needed
	up := false
	for i, op := range ops {
		// if FEN is missing, convert sequence to FEN
		if len(op.SequenceFENs) == 0 {
			op.SequenceFENs = convertToFEN(op.Moves)
			op.ResultFEN = op.SequenceFENs[len(op.SequenceFENs)-1]
			fmt.Printf("Converted sequence in opening to FEN: %s, Sequence: %s, FEN: %s\n", op.Name, op.Moves, op.ResultFEN)
			ops[i] = op
			up = true
		}
	}

	if up {
		if err := writeOpenings(ops); err != nil {
			fmt.Printf("Error writing openings: %v\n", err)
			return
		}

		fmt.Printf("Updated openings file: %s\n", openingsCSV)
	}

	fmt.Printf("Processed %d openings\n", len(ops))
}
