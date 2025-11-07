package chess

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func createTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "openings.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp csv: %v", err)
	}
	return path
}

func TestCreateOpeningsLibrary(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		csvContent := `Moves,Name,ResultFEN,SequenceFENs
"1. e4 e5",Ruy Lopez,fen3,"fen1,fen2,fen3"`
		path := createTempCSV(t, csvContent)

		ol := &openingsLibrary{}
		err := ol.readOpenings(path)
		if err != nil {
			t.Fatalf("readOpenings() error = %v, wantErr nil", err)
		}

		if len(ol.ops) != 1 {
			t.Fatalf("expected 1 opening, got %d", len(ol.ops))
		}

		expected := Opening{
			Moves:        []string{"e4", "e5"},
			Name:         "Ruy Lopez",
			ResultFEN:    "fen3",
			SequenceFENs: []string{"fen1", "fen2", "fen3"},
		}

		if !reflect.DeepEqual(ol.ops[0], expected) {
			t.Errorf("opening mismatch:\ngot:  %+v\nwant: %+v", ol.ops[0], expected)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		ol := &openingsLibrary{}
		err := ol.readOpenings("nonexistent.csv")
		if err == nil {
			t.Fatal("expected an error for a non-existent file, got nil")
		}
	})
}

func TestFindOpening(t *testing.T) {
	csvContent := `Moves,Name,ResultFEN,SequenceFENs
"1. e4 e5",Ruy Lopez,fen3,"fen1,fen2,fen3"
"1. d4 d5",Queen's Gambit,qg3,"qg1,qg2,qg3"`
	path := createTempCSV(t, csvContent)

	ol := &openingsLibrary{}
	if err := ol.readOpenings(path); err != nil {
		t.Fatalf("readOpenings() failed: %v", err)
	}

	t.Run("find existing opening", func(t *testing.T) {
		op, found := ol.FindOpening("fen3")
		if !found {
			t.Fatal("expected to find an opening, but did not")
		}
		if op.Name != "Ruy Lopez" {
			t.Errorf("expected opening 'Ruy Lopez', got '%s'", op.Name)
		}
	})

	t.Run("do not find non-existent opening", func(t *testing.T) {
		_, found := ol.FindOpening("nonexistent_fen")
		if found {
			t.Fatal("expected not to find an opening, but did")
		}
	})
}

func TestFindVariations(t *testing.T) {
	csvContent := `Moves,Name,ResultFEN,SequenceFENs
"1. e4 e5 2. Nf3",Ruy Lopez Variation,fen4,"fen1,fen2,fen3,fen4"
"1. e4 c5",Sicilian Defence,sic2,"fen1,sic2"
"1. d4 d5",Queen's Gambit,qg3,"qg1,qg2,qg3"`
	path := createTempCSV(t, csvContent)

	ol := &openingsLibrary{}
	if err := ol.readOpenings(path); err != nil {
		t.Fatalf("readOpenings() failed: %v", err)
	}

	t.Run("find variations for a common FEN", func(t *testing.T) {
		variations, found := ol.FindVariations("fen1")
		if !found {
			t.Fatal("expected to find variations, but did not")
		}

		expected := []string{"fen1", "fen2", "fen3"}
		slices.Sort(variations)
		slices.Sort(expected)

		if !reflect.DeepEqual(variations, expected) {
			t.Errorf("variation mismatch:\ngot:  %v\nwant: %v", variations, expected)
		}
	})

	t.Run("find variations for a less common FEN", func(t *testing.T) {
		variations, found := ol.FindVariations("fen2")
		if !found {
			t.Fatal("expected to find variations, but did not")
		}

		expected := []string{"fen2", "fen3"}
		slices.Sort(variations)
		slices.Sort(expected)

		if !reflect.DeepEqual(variations, expected) {
			t.Errorf("variation mismatch:\ngot:  %v\nwant: %v", variations, expected)
		}
	})

	t.Run("do not find variations for a FEN not in a sequence", func(t *testing.T) {
		_, found := ol.FindVariations("nonexistent_fen")
		if found {
			t.Fatal("expected not to find variations, but did")
		}
	})

	t.Run("do not find variations for a result FEN", func(t *testing.T) {
		// `FindVariations` should only find FENs that are *intermediate* steps
		_, found := ol.FindVariations("fen4")
		if found {
			t.Fatal("expected not to find variations for a result FEN, but did")
		}
	})

	t.Run("uniqueness of variations", func(t *testing.T) {
		csvContent := `Moves,Name,ResultFEN,SequenceFENs
"1. e4 e5",Opening 1,fen3,"fen1,fen2,fen3"
"1. e4 c5",Opening 2,fen4,"fen1,fen2,fen4"`
		path := createTempCSV(t, csvContent)

		ol := &openingsLibrary{}
		if err := ol.readOpenings(path); err != nil {
			t.Fatalf("readOpenings() failed: %v", err)
		}

		variations, found := ol.FindVariations("fen1")
		if !found {
			t.Fatal("expected to find variations, but did not")
		}

		expected := []string{"fen1", "fen2"}
		slices.Sort(variations)
		slices.Sort(expected)

		if !reflect.DeepEqual(variations, expected) {
			t.Errorf("variation mismatch:\ngot:  %v\nwant: %v", variations, expected)
		}
	})
}
