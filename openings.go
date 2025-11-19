// Package chess provides a chess engine with a focus on algebraic notation,
// game state management, and move validation.
package chess

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

//go:embed data/openings.csv
var embeddedOpeningsCSV []byte

// Opening represents a single chess opening, including its name, move sequence,
// and the FEN strings for each position in the sequence.
type Opening struct {
	// ECO is the Encyclopedia of Chess Openings code (e.g., "A00").
	ECO string
	// Moves is a slice of algebraic notation strings representing the move sequence.
	Moves []string
	// Name is the common name of the opening (e.g., "Ruy Lopez").
	Name string
	// ResultFEN is the Forsyth-Edwards Notation string of the final board position.
	ResultFEN string
	// SequenceFENs is a slice of FEN strings for each board state in the opening sequence.
	SequenceFENs []string
	// SequenceMoves is a string representation of each turn
	SequenceMoves string
}

// openingsLibrary holds a collection of chess openings loaded from a data source.
type openingsLibrary struct {
	ops []Opening
}

// CreateOpeningsLibrary initializes a new openings library by reading from the
// default openings CSV file. It returns an error if the file cannot be read or parsed.
func CreateOpeningsLibrary() (*openingsLibrary, error) {
	ol := &openingsLibrary{}

	// load openings from the embedded CSV so downstream consumers do not have
	// to keep the data file on disk.
	if err := ol.loadOpenings(bytes.NewReader(embeddedOpeningsCSV)); err != nil {
		return nil, fmt.Errorf("error reading openings: %w", err)
	}

	return ol, nil
}

// readOpenings reads and parses opening data from a CSV file at the given path.
func (ol *openingsLibrary) readOpenings(pth string) error {
	f, err := os.Open(pth)
	if err != nil {
		return err
	}
	defer f.Close()

	return ol.loadOpenings(f)
}

// loadOpenings converts CSV data from the provided reader into the in-memory
// openings library representation.
func (ol *openingsLibrary) loadOpenings(r io.Reader) error {
	reader := csv.NewReader(r)
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

		// if we're missing fields, move on
		if len(record) != 5 {
			continue
		}

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
			Moves:         mvs,
			ECO:           record[1],
			Name:          record[2],
			ResultFEN:     record[3],
			SequenceFENs:  strings.Split(record[4], ","),
			SequenceMoves: record[0],
		})
	}

	// update library with processed openings
	ol.ops = ops

	// process openings as needed
	return nil
}

// All returns an iterator
func (ol *openingsLibrary) All() func(yld func(Opening) bool) {
	return func(yld func(Opening) bool) {
		for _, op := range ol.ops {
			if !yld(op) {
				break
			}
		}
	}
}

// FindOpeningsByECO searches the library for an opening that is classified by the
// provided ECO string. It returns an opening if found, and a boolean indicating success.
func (ol *openingsLibrary) FindOpeningByECO(eco string) (*Opening, bool) {
	for _, op := range ol.ops {
		if op.ECO == eco {
			return &op, true
		}
	}

	return nil, false
}

// FindOpeningByFEN searches the library for an opening that results in the given FEN string.
// It returns the opening if found, and a boolean indicating success.
func (ol *openingsLibrary) FindOpeningByFEN(fen string) (*Opening, bool) {
	for _, op := range ol.ops {
		if op.ResultFEN == fen {
			return &op, true
		}
	}

	return nil, false
}

// FindVariationsByFEN searches for all known opening sequences that include the given FEN.
// It returns a slice of unique subsequent FENs from all matching sequences,
// representing possible continuations. The boolean indicates if any variations were found.
func (ol *openingsLibrary) FindVariationsByFEN(fen string) ([]string, bool) {
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
