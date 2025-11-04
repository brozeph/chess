package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	chess "github.com/brozeph/chess"
)

func printBoard(b *chess.Board) {
	fmt.Println("Current board position:")
	for rank := 8; rank >= 1; rank-- {
		fmt.Printf("%d ", rank)
		for file := 'a'; file <= 'h'; file++ {
			sq := b.GetSquare(file, rank)
			fmt.Printf("%c ", sq.Piece.AlgebraicSymbol())
		}
		fmt.Println()
	}
	fmt.Println("  a b c d e f g h")
}

func main() {
	var fen string
	var movesArg string
	var usePGN bool

	flag.StringVar(&fen, "fen", "", "starting FEN (defaults to initial position)")
	flag.StringVar(&movesArg, "moves", "", "space separated algebraic moves to apply")
	flag.BoolVar(&usePGN, "pgn", false, "interpret castling notation using PGN (O-O)")
	flag.Parse()

	opts := chess.AlgebraicClientOptions{PGN: usePGN}

	var client *chess.AlgebraicGameClient
	var err error

	if strings.TrimSpace(fen) != "" {
		client, err = chess.CreateAlgebraicGameClientFromFEN(fen, opts)
		if err != nil {
			log.Fatalf("failed to load FEN: %v", err)
		}
	} else {
		client = chess.CreateAlgebraicGameClient(opts)
	}

	if movesArg != "" {
		for _, mv := range strings.Fields(movesArg) {
			if _, err := client.Move(mv, false); err != nil {
				log.Fatalf("failed to apply move %s: %v", mv, err)
			}
		}
	}

	status, err := client.Status(false)
	if err != nil {
		log.Fatalf("failed to evaluate position: %v", err)
	}

	printBoard(status.Board)
	fmt.Printf("FEN: %s\n", client.FEN())
	fmt.Printf("Check: %t  Checkmate: %t  Stalemate: %t  Repetition: %t\n",
		status.IsCheck, status.IsCheckmate, status.IsStalemate, status.IsRepetition)

	keys := make([]string, 0, len(status.NotatedMoves))
	for k := range status.NotatedMoves {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("Next moves (%d):\n", len(keys))
	for _, k := range keys {
		fmt.Println(" -", k)
	}
}
