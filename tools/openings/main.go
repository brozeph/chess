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
	FEN      string
	Name     string
	Sequence string
}

func convertToFEN(seq string) string {
	c := chess.CreateAlgebraicGameClient()

	// break the sequence into moves (format is 1 e4 e5 2 Nf3 Nc6 ...)
	prts := strings.Fields(seq)

	for i, prt := range prts {
		// skip move numbers (the first in a sequence of 3 parts)
		if i%3 == 0 {
			fmt.Print("processing move number:", prt, "\n")
			continue
		}

		// process moves here to build FEN (not implemented)
		c.Move(prt)
	}

	return c.FEN()
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

func writeOpenings(ops []opening) error {
	// write updated openings to CSV file
	f, err := os.Create(openingsCSV)
	if err != nil {
		return fmt.Errorf("error creating openings file: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	for _, op := range ops {
		err := writer.Write([]string{op.FEN, op.Sequence, op.Name})
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
		if op.FEN == "" {
			op.FEN = convertToFEN(op.Sequence)
			fmt.Printf("Converted sequence in opening to FEN: %s, Sequence: %s, FEN: %s\n", op.Name, op.Sequence, op.FEN)
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
