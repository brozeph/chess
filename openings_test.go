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
		csvContent := `Moves,ECO,Name,ResultFEN,SequenceFENs
"1. e4 e5",C60,Ruy Lopez,fen3,"fen1,fen2,fen3"`
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
			Moves:         []string{"e4", "e5"},
			ECO:           "C60",
			Name:          "Ruy Lopez",
			ResultFEN:     "fen3",
			SequenceFENs:  []string{"fen1", "fen2", "fen3"},
			SequenceMoves: "1. e4 e5",
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

func TestFindOpeningByFEN(t *testing.T) {
	csvContent := `Moves,ECO,Name,ResultFEN,SequenceFENs
"1. e4 e5",C60,Ruy Lopez,fen3,"fen1,fen2,fen3"
"1. d4 d5",D00,Queen's Gambit,qg3,"qg1,qg2,qg3"`
	path := createTempCSV(t, csvContent)

	ol := &openingsLibrary{}
	if err := ol.readOpenings(path); err != nil {
		t.Fatalf("readOpenings() failed: %v", err)
	}

	t.Run("find existing opening", func(t *testing.T) {
		op, found := ol.FindOpeningByFEN("fen3")
		if !found {
			t.Fatal("expected to find an opening, but did not")
		}
		if op.Name != "Ruy Lopez" {
			t.Errorf("expected opening 'Ruy Lopez', got '%s'", op.Name)
		}
	})

	t.Run("do not find non-existent opening", func(t *testing.T) {
		_, found := ol.FindOpeningByFEN("nonexistent_fen")
		if found {
			t.Fatal("expected not to find an opening, but did")
		}
	})
}

func TestFindVariationsByFEN(t *testing.T) {
	csvContent := `Moves,ECO,Name,ResultFEN,SequenceFENs
"1. e4 e5 2. Nf3",C60,Ruy Lopez Variation,fen4,"fen1,fen2,fen3,fen4"
"1. e4 c5",B20,Sicilian Defence,sic2,"fen1,sic2"
"1. d4 d5",D00,Queen's Gambit,qg3,"qg1,qg2,qg3"`
	path := createTempCSV(t, csvContent)

	ol := &openingsLibrary{}
	if err := ol.readOpenings(path); err != nil {
		t.Fatalf("readOpenings() failed: %v", err)
	}

	t.Run("find variations for a common FEN", func(t *testing.T) {
		variations, found := ol.FindVariationsByFEN("fen1")
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
		variations, found := ol.FindVariationsByFEN("fen2")
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
		_, found := ol.FindVariationsByFEN("nonexistent_fen")
		if found {
			t.Fatal("expected not to find variations, but did")
		}
	})

	t.Run("do not find variations for a result FEN", func(t *testing.T) {
		// `FindVariations` should only find FENs that are *intermediate* steps
		_, found := ol.FindVariationsByFEN("fen4")
		if found {
			t.Fatal("expected not to find variations for a result FEN, but did")
		}
	})

	t.Run("uniqueness of variations", func(t *testing.T) {
		csvContent := `Moves,ECO,Name,ResultFEN,SequenceFENs
"1. e4 e5",C60,Opening 1,fen3,"fen1,fen2,fen3"
"1. e4 c5",B20,Opening 2,fen4,"fen1,fen2,fen4"`
		path := createTempCSV(t, csvContent)

		ol := &openingsLibrary{}
		if err := ol.readOpenings(path); err != nil {
			t.Fatalf("readOpenings() failed: %v", err)
		}

		variations, found := ol.FindVariationsByFEN("fen1")
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

func TestOpeningsLibraryAll(t *testing.T) {
	ops := []Opening{
		{ECO: "A00", Name: "Opening 1"},
		{ECO: "A01", Name: "Opening 2"},
		{ECO: "A02", Name: "Opening 3"},
	}

	ol := &openingsLibrary{ops: ops}

	t.Run("iterates all openings", func(t *testing.T) {
		iter := ol.All()
		var names []string
		iter(func(op Opening) bool {
			names = append(names, op.Name)
			return true
		})

		if !reflect.DeepEqual(names, []string{"Opening 1", "Opening 2", "Opening 3"}) {
			t.Fatalf("unexpected iteration order: %v", names)
		}
	})

	t.Run("stops when yield returns false", func(t *testing.T) {
		iter := ol.All()
		count := 0
		iter(func(op Opening) bool {
			count++
			return count < 2
		})

		if count != 2 {
			t.Fatalf("expected early termination after 2 yields, got %d", count)
		}
	})
}

func TestFindOpeningByECO(t *testing.T) {
	csvContent := `Moves,ECO,Name,ResultFEN,SequenceFENs
"1. e4 e5",C60,Ruy Lopez,fen3,"fen1,fen2,fen3"
"1. d4 d5",D00,Queen's Gambit,qg3,"qg1,qg2,qg3"`
	path := createTempCSV(t, csvContent)

	ol := &openingsLibrary{}
	if err := ol.readOpenings(path); err != nil {
		t.Fatalf("readOpenings() failed: %v", err)
	}

	t.Run("find existing ECO", func(t *testing.T) {
		op, found := ol.FindOpeningByECO("C60")
		if !found {
			t.Fatal("expected to find opening by ECO")
		}
		if op.Name != "Ruy Lopez" {
			t.Fatalf("expected Ruy Lopez, got %s", op.Name)
		}
	})

	t.Run("missing ECO", func(t *testing.T) {
		if _, found := ol.FindOpeningByECO("C99"); found {
			t.Fatal("expected not to find opening for missing ECO")
		}
	})
}
