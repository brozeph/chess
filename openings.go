// Package chess provides a chess engine with a focus on algebraic notation,
// game state management, and move validation.
package chess

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	openingsCSV = "./data/openings.csv"
)

// Opening represents a single chess opening, including its name, move sequence,
// and the FEN strings for each position in the sequence.
type Opening struct {
	// Moves is a slice of algebraic notation strings representing the move sequence.
	Moves []string
	// Name is the common name of the opening (e.g., "Ruy Lopez").
	Name string
	// ResultFEN is the Forsyth-Edwards Notation string of the final board position.
	ResultFEN string
	// SequenceFENs is a slice of FEN strings for each board state in the opening sequence.
	SequenceFENs []string
}

// openingsLibrary holds a collection of chess openings loaded from a data source.
type openingsLibrary struct {
	ops []Opening
}

// CreateOpeningsLibrary initializes a new openings library by reading from the
// default openings CSV file. It returns an error if the file cannot be read or parsed.
func CreateOpeningsLibrary() (*openingsLibrary, error) {
	ol := &openingsLibrary{}

	// read in openings from CSV file
	if err := ol.readOpenings(openingsCSV); err != nil {
		return nil, fmt.Errorf("error reading openings: %w", err)
	}

	return ol, nil
}

// readOpenings reads and parses opening data from a CSV file at the given path.
func (ol *openingsLibrary) readOpenings(pth string) error {
	// read in openings from CSV file
	f, err := os.Open(pth)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	ops := []Opening{}
	for i, record := range records {
		// skip header row
		if i == 0 {
			continue
		}

		// only read valid rows
		if len(record) == 4 {
			mvs := []string{}
			for i, trn := range strings.Fields(record[0]) {
				// the turn number
				if i%3 == 0 {
					continue
				}

				mvs = append(mvs, trn)
			}

			// create opening from CSV data
			ops = append(ops, Opening{
				Moves:        mvs,
				Name:         record[1],
				ResultFEN:    record[2],
				SequenceFENs: strings.Split(record[3], ","),
			})

			continue
		}
	}

	// update library with processed openings
	ol.ops = ops

	// process openings as needed
	return nil
}

// FindOpening searches the library for an opening that results in the given FEN string.
// It returns the opening if found, and a boolean indicating success.
func (ol *openingsLibrary) FindOpening(fen string) (*Opening, bool) {
	for _, op := range ol.ops {
		if op.ResultFEN == fen {
			return &op, true
		}
	}

	return nil, false
}

// FindVariations searches for all known opening sequences that include the given FEN.
// It returns a slice of unique subsequent FENs from all matching sequences,
// representing possible continuations. The boolean indicates if any variations were found.
func (ol *openingsLibrary) FindVariations(fen string) ([]string, bool) {
	mtchs := []string{}

	for _, op := range ol.ops {
		l := len(op.SequenceFENs)
		// look for openings where the fen is part of the sequence
		if l > 1 {
			if i := slices.Index(op.SequenceFENs[:l-1], fen); i >= 0 {
				// track each fen in the sequence from the match forward
				for _, s := range op.SequenceFENs[i : l-1] {
					// store unique FENs
					if slices.Contains(mtchs, s) {
						continue
					}

					mtchs = append(mtchs, s)
				}
			}
		}
	}

	return mtchs, len(mtchs) > 0
}
