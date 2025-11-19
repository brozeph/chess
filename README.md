# Chess (Go Edition)

A Go port of the [Node Chess](https://brozeph.github.io/node-chess/) engine that focuses on algebraic notation.  
It parses and validates moves, tracks rich game state, and exposes an event-driven API that integrates cleanly with Go applications.

## Featuring

- **Notation-first game play** – list every legal move in algebraic notation, accepts algebraic input and surface promotion choices.
- **Robust state inspection** – detect check, checkmate, stalemate, and threefold repetition while keeping a complete capture and move history.
- **Undo-friendly move execution** – every applied move returns an undo handle and updates castling rights, en passant targets, and move counters.
- **FEN integration** – load games from Forsyth–Edwards Notation, emit FEN snapshots after every move, or explore alternate continuations.
- **Event-driven hooks** – subscribe to move, capture, castle, promotion, undo, check, and checkmate notifications from multiple abstraction layers.
- **Opening library** - iterable and searchable library of openings, with ECO and FEN

## Table of Contents

- [Installation](#installation)
- [Quickstart](#quickstart)
- [Inspecting Valid Moves](#inspecting-valid-moves)
- [Making and Undoing Moves](#making-and-undoing-moves)
- [Loading Custom Positions](#loading-custom-positions)
- [Event API](#event-api)
- [Understanding Returned Types](#understanding-returned-types)
- [CLI Example](#cli-example)
- [Development](#development)
- [License](#license)

## Installation

```bash
go get github.com/brozeph/chess@latest
```

Import the package in your Go code:

```go
import "github.com/brozeph/chess"
```

## Quickstart

```go
package main

import (
 "fmt"
 "log"

 "github.com/brozeph/chess"
)

func main() {
 client := chess.CreateAlgebraicGameClient()

 status, err := client.Status()
 if err != nil {
  log.Fatal(err)
 }

 fmt.Printf("FEN: %s\n", client.FEN())
 fmt.Printf("Side to move: %s\n", status.Side().Name())

 for notation := range status.NotatedMoves {
  fmt.Println(notation)
 }
}
```

## Using the Opening Library

The package ships with a curated ECO-backed opening database (`data/openings.csv`). Load it once and iterate or search:

```go
ol, err := chess.CreateOpeningsLibrary()
if err != nil {
 log.Fatal(err)
}

// Iterate all openings (10k+ entries) lazily
iter := ol.All()
iter(func(op chess.Opening) bool {
 fmt.Printf("%s %s -> %s\n", op.ECO, op.Name, op.ResultFEN)
 return true
})

// Look up by ECO or by final FEN
if op, ok := ol.FindOpeningByECO("C60"); ok {
 fmt.Println("Found:", op.Name)
}

if op, ok := ol.FindOpeningByFEN("r1bqkbnr/pppp1ppp/2n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 3 3"); ok {
 fmt.Println("Reached via:", op.SequenceMoves)
}

// Explore continuations after a position appears anywhere in the sequence.
if next, ok := ol.FindVariationsByFEN("r1bqkbnr/pppp1ppp/2n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 3 3"); ok {
 fmt.Println("Possible continuations:", next)
}
```

Each `Opening` exposes the ECO code, friendly name, raw move text (`SequenceMoves`), and every intermediate FEN (`SequenceFENs`), making it easy to link engine positions back to common theory.

## Inspecting Valid Moves

`Status()` returns a `*chess.GameStatus` that stays in sync with the underlying game:

```go
status, err := client.Status()
if err != nil {
 log.Fatal(err)
}

fmt.Println("Check:", status.IsCheck)
fmt.Println("Checkmate:", status.IsCheckmate)
fmt.Println("Stalemate:", status.IsStalemate)
fmt.Println("Threefold repetition:", status.IsRepetition)

for algebraic, move := range status.NotatedMoves {
 src := fmt.Sprintf("%c%d", move.Src.File, move.Src.Rank)
 dst := fmt.Sprintf("%c%d", move.Dest.File, move.Dest.Rank)
 fmt.Printf("%-5s -> %s -> %s\n", algebraic, src, dst)
}
```

The `NotatedMoves` map is keyed by algebraic notation and each entry exposes the source/destination squares through `move.Src` and `move.Dest`.

## Making and Undoing Moves

```go
result, err := client.Move("e4")
if err != nil {
 log.Fatalf("illegal move: %v", err)
}

fmt.Printf("Played: %s\n", result.Move.Algebraic)
fmt.Printf("Piece: %s\n", result.Move.Piece.Notation)
fmt.Printf("FEN after move: %s\n", client.FEN())

// Undo later if needed.
result.Undo()
```

- Promotions can be specified by suffixing the desired piece (`e8=Q`, `exd8N`, etc.).  
- `result.Move` gives full context including castling, en passant, captured piece, and rook movement when appropriate.

## Loading Custom Positions

```go
client, err := chess.CreateAlgebraicGameClientFromFEN(
 "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
)
if err != nil {
 log.Fatal(err)
}

status, _ := client.Status()
fmt.Printf("Side to move: %s\n", status.Side().Name())
```

- All FEN components (castling availability, en-passant target, half-move clock, full-move number) are respected.  
- Call `Status(true)` to force recalculation if you have manipulated the underlying board state directly.

## Event API

Subscribe to events using `On`:

```go
client.On("move", func(data interface{}) {
 if mv, ok := data.(*chess.MoveEvent); ok {
  fmt.Printf("%s to %c%d\n", mv.Algebraic, mv.PostSquare.File, mv.PostSquare.Rank)
 }
})

client.On("checkmate", func(data interface{}) {
 if side, ok := data.(chess.Side); ok {
  fmt.Printf("%s was checkmated\n", side.Name())
 }
})
```

| Event       | Payload type           | Description |
|-------------|------------------------|-------------|
| `move`      | `*chess.MoveEvent`     | Emitted after every legal move. |
| `capture`   | `*chess.MoveEvent`     | Fired when a capture occurs. |
| `castle`    | `*chess.MoveEvent`     | Fired after a king-side or queen-side castle. |
| `enPassant` | `*chess.MoveEvent`     | Fired when an en passant capture is performed. |
| `promote`   | `*chess.Square`        | Triggered after a pawn promotion; the square contains the promoted piece. |
| `undo`      | `*chess.MoveEvent`     | Emitted after a move has been reverted. |
| `check`     | `chess.Side`           | Indicates which side is currently in check. |
| `checkmate` | `chess.Side`           | Indicates which side has been checkmated. |

Events propagate from the board to the game and up to the algebraic client, so you can subscribe at whichever layer you interact with.

## Understanding Returned Types

- `*chess.GameStatus` – encapsulates the current `Game`, flags for check/checkmate/stalemate/repetition, and a `NotatedMoves` map. Use `status.Side().Name()` to see whose turn it is (or `status.Side().Opponent()` to get the opposing player).  
- `*chess.MoveEvent` – describes the move that just executed. Access prior and post squares (`PrevSquare`, `PostSquare`), captured pieces, promotion metadata, and rook movement on castling.  
- `*chess.Square` – exposes `File`, `Rank`, and the occupying `*chess.Piece`.  
- `*chess.Piece` – contains `Type`, `Side`, `Notation`, and `MoveCount`.  
- `notationMove` entries – each value from `status.NotatedMoves` exposes `Src` and `Dest` squares and a `FEN(fen string)` helper that applies the move to an arbitrary position.

## CLI Example

The repository ships with a small CLI that mirrors the quickstart above:

```bash
go run ./examples/main.go \
  -fen "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1" \
  -moves "Nc6 Nc3"
```

Flags:

- `-fen` – optional starting FEN (defaults to the initial position).  
- `-moves` – space separated algebraic moves to apply before printing the board.  
- `-pgn` – treat castling notation as PGN (`O-O`) instead of numeric (`0-0`).

Running without flags prints the initial position and the 20 legal opening moves.

## Development

- Go version: see `go.mod` (currently Go 1.24).  
- Run tests:

  ```bash
  go test ./...
  ```

- Format code with `gofmt` before submitting patches.

## License

MIT © Joshua Thomas. See [LICENSE](LICENSE) for details.
