package chess

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func getValidMovesByPieceType(pt pieceType, validMoves []potentialMoves) []potentialMoves {
	res := []potentialMoves{}
	for _, mv := range validMoves {
		if mv.origin.Piece != nil && mv.origin.Piece.Type == pt {
			res = append(res, mv)
		}
	}
	return res
}

func getNotationPrefix(src *Square, dest *Square, moves []potentialMoves) string {
	prefix := src.Piece.Notation
	fileCount := map[rune]int{}
	rankCount := map[int]int{}

	for _, mv := range moves {
		for _, sq := range mv.destinationSquares {
			if sq == dest {
				fileCount[mv.origin.File]++
				rankCount[mv.origin.Rank]++
			}
		}
	}

	if len(fileCount) > 1 {
		prefix += string(src.File)
	}

	if len(rankCount) > len(fileCount) {
		prefix += strconv.Itoa(src.Rank)
	}

	return prefix
}

func sanitizeNotation(n string, usePGN bool) string {
	clean := strings.ReplaceAll(n, "!", "")
	clean = strings.ReplaceAll(clean, "+", "")
	clean = strings.ReplaceAll(clean, "#", "")
	clean = strings.ReplaceAll(clean, "=", "")
	clean = strings.ReplaceAll(clean, "\\", "")

	if usePGN {
		clean = strings.ReplaceAll(clean, "0", "O")
		return clean
	}

	clean = strings.ReplaceAll(clean, "O", "0")
	return clean
}

// AlgebraicClientOptions provides configuration options for an AlgebraicGameClient.
type AlgebraicClientOptions struct {
	PGN bool // PGN specifies whether to use PGN-style notation for castling (O-O) instead of (0-0).
}

// AlgebraicGameClient provides a client for interacting with a chess game using algebraic notation.
type AlgebraicGameClient struct {
	fen          string
	game         *Game
	isCheck      bool
	isCheckmate  bool
	isRepetition bool
	isStalemate  bool
	notatedMoves map[string]notationMove
	options      AlgebraicClientOptions
	validMoves   []potentialMoves
	validation   *gameValidator

	events *eventHub
}

// CreateAlgebraicGameClient creates a new game client with a standard starting board.
// It accepts optional AlgebraicClientOptions.
func CreateAlgebraicGameClient(opts ...AlgebraicClientOptions) *AlgebraicGameClient {
	var o AlgebraicClientOptions
	if len(opts) > 0 {
		o = opts[0]
	}

	g := createGame()
	client := &AlgebraicGameClient{
		game:         g,
		notatedMoves: map[string]notationMove{},
		options:      o,
		validation:   CreateGameValidator(g),
		validMoves:   []potentialMoves{},
		events:       newEventHub(),
	}
	client.bindGameEvents()
	client.On("undo", func(interface{}) {
		_ = client.update()
	})
	_ = client.update()
	return client
}

// CreateAlgebraicGameClientFromFEN creates a new game client from a FEN string.
// It returns an error if the FEN string is invalid.
func CreateAlgebraicGameClientFromFEN(fen string, opts ...AlgebraicClientOptions) (*AlgebraicGameClient, error) {
	if strings.TrimSpace(fen) == "" {
		return nil, errors.New("FEN must be a non-empty string")
	}

	var o AlgebraicClientOptions
	if len(opts) > 0 {
		o = opts[0]
	}

	// load the board state (piece positions) from the FEN
	loaded, err := loadBoard(fen)
	if err != nil {
		return nil, err
	}

	// process the remainder of the FEN string
	parts := strings.Split(fen, " ")

	// check the active color
	active := "w"
	if len(parts) > 1 {
		active = parts[1]
	}

	bs := sideWhite
	if active == "b" {
		bs = sideBlack
	}

	// create a game container for history and event tracking
	g := createGame(bs == sideWhite)
	g.Board = loaded
	g.Board.LastMovedPiece = nil
	g.hookBoardEvents()

	// track castling availability
	if len(parts) > 2 {
		g.cstl = parts[2]
	}

	// track en-passant square
	if len(parts) > 3 {
		g.enP = g.Board.getSquareByName(parts[3])
	}

	// track halfmove clock
	if len(parts) > 4 {
		g.hmc, _ = strconv.Atoi(parts[4])
	}

	// track fullmove number
	if len(parts) > 5 {
		g.fmn, _ = strconv.Atoi(parts[5])
	}

	client := &AlgebraicGameClient{
		fen:          fen,
		game:         g,
		notatedMoves: map[string]notationMove{},
		options:      o,
		validation:   CreateGameValidator(g),
		validMoves:   []potentialMoves{},
		events:       newEventHub(),
	}

	client.bindGameEvents()
	client.On("undo", func(interface{}) {
		_ = client.update()
	})

	if err := client.update(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *AlgebraicGameClient) bindGameEvents() {
	if c.game == nil {
		return
	}

	bubble := func(event string) {
		c.game.on(event, func(data interface{}) {
			c.emit(event, data)
		})
	}

	for _, ev := range []string{"move", "capture", "castle", "enPassant", "promote"} {
		bubble(ev)
	}

	c.game.on("undo", func(data interface{}) {
		c.emit("undo", data)
	})

	c.game.on("check", func(data interface{}) {
		c.emit("check", data)
	})

	c.game.on("checkmate", func(data interface{}) {
		c.emit("checkmate", data)
	})
}

func (c *AlgebraicGameClient) emit(ev string, d any) {
	if c == nil {
		return
	}

	c.events.emit(ev, d)
}

func (c *AlgebraicGameClient) notate(mvs []potentialMoves) map[string]notationMove {
	algebraic := map[string]notationMove{}

	for _, vm := range mvs {
		src := vm.origin
		if src.Piece == nil {
			continue
		}
		for _, dest := range vm.destinationSquares {
			prefix := ""
			suffix := ""
			isPromotion := false

			if dest.Piece != nil {
				suffix = "x"
			}
			suffix += dest.name()

			if dest.Rank == 1 || dest.Rank == 8 {
				isPromotion = src.Piece.Type == piecePawn
			}

			if dest.Piece != nil && src.Piece.Type == piecePawn {
				prefix = string(src.File)
			}

			if src.Piece.Type == piecePawn &&
				src.File != dest.File &&
				dest.Piece == nil {
				prefix = string(src.File) + "x"
			}

			switch src.Piece.Type {
			case pieceBishop, pieceKnight, pieceQueen, pieceRook:
				matches := getValidMovesByPieceType(src.Piece.Type, mvs)
				prefix = src.Piece.Notation
				if len(matches) > 1 {
					prefix = getNotationPrefix(src, dest, matches)
				}
			case pieceKing:
				prefix = src.Piece.Notation
				if src.File == 'e' && dest.File == 'g' {
					prefix = "0-0"
					if c.options.PGN {
						prefix = "O-O"
					}
					suffix = ""
				}
				if src.File == 'e' && dest.File == 'c' {
					prefix = "0-0-0"
					if c.options.PGN {
						prefix = "O-O-O"
					}
					suffix = ""
				}
			case piecePawn:
				if prefix == "" && dest.Piece == nil {
					prefix = ""
				}
			default:
				if prefix == "" {
					prefix = src.Piece.Notation
				}
			}

			if prefix == "" && src.Piece.Type != piecePawn {
				prefix = src.Piece.Notation
			}

			if !isPromotion {
				key := prefix + suffix
				algebraic[key] = notationMove{Src: src, Dest: dest}
				continue
			}

			for _, promo := range []string{"R", "N", "B", "Q"} {
				key := prefix + suffix + promo
				algebraic[key] = notationMove{Src: src, Dest: dest}
			}
		}
	}

	return algebraic
}

func (c *AlgebraicGameClient) update() error {
	result, err := c.validation.Check()
	if err != nil {
		return err
	}
	c.isCheck = result.IsCheck
	c.isCheckmate = result.IsCheckmate
	c.isRepetition = result.IsRepetition
	c.isStalemate = result.IsStalemate
	c.validMoves = result.ValidMoves
	c.notatedMoves = c.notate(result.ValidMoves)
	return nil
}

// CaptureHistory returns a slice of pieces that have been captured during the game.
func (c *AlgebraicGameClient) CaptureHistory() []*Piece {
	return c.game.CaptureHistory
}

// FEN returns the Forsyth-Edwards Notation (FEN) string for the current board state.
func (c *AlgebraicGameClient) FEN() string {
	return c.game.fen()
}

// Move attempts to make a move using algebraic notation.
func (c *AlgebraicGameClient) Move(ntn string) (*moveResult, error) {
	if ntn == "" {
		return nil, errors.New("notation is invalid")
	}

	origNtn := ntn
	ntn = sanitizeNotation(ntn, c.options.PGN)

	var prmP string
	if len(ntn) > 0 {
		p := ntn[len(ntn)-1]
		if strings.ContainsRune("BNQR", rune(p)) {
			prmP = string(p)
		}
	}

	// Fallback for verbose notations like "Nb1c3" when "Nc3" is expected.
	// If the direct lookup fails, try to parse it.
	if _, ok := c.notatedMoves[ntn]; !ok && len(ntn) >= 4 {
		// A piece notation (e.g., N, B, R, Q, K) is one character.
		// A pawn move would be something like e2e4, which is also 4 characters.
		p := ""
		ofs := 0
		if strings.ContainsRune("NBRQK", rune(ntn[0])) {
			p = string(ntn[0])
			ofs = 1
		}

		if len(ntn) >= ofs+4 {
			srcN := ntn[ofs : ofs+2]
			dstN := ntn[ofs+2 : ofs+4]

			// Find the corresponding standard notation move
			for k, mv := range c.notatedMoves {
				if mv.Src.name() == srcN && mv.Dest.name() == dstN && strings.HasPrefix(k, p) {
					ntn = k // Found it, use the standard notation
					break
				}
			}
		}
	}

	if mv, ok := c.notatedMoves[ntn]; ok {
		res, err := c.game.move(mv.Src, mv.Dest, ntn)
		if err != nil {
			return nil, err
		}

		if prmP != "" {
			var p *Piece
			side := c.game.getCurrentSide().Opponent()
			switch prmP {
			case "B":
				p = newPiece(pieceBishop, side)
			case "N":
				p = newPiece(pieceKnight, side)
			case "Q":
				p = newPiece(pieceQueen, side)
			case "R":
				p = newPiece(pieceRook, side)
			}
			if p != nil {
				if _, err := c.game.promote(res.Move.PostSquare, p); err != nil {
					return nil, err
				}
			}
		}

		if err := c.update(); err != nil {
			return nil, err
		}

		return res, nil
	}

	return nil, fmt.Errorf("notation is invalid (%s)", origNtn)
}

// On registers an event handler for the given event.
// The client supports the following events:
//   - "move":      emitted after a piece has been moved. The handler receives a *MoveEvent.
//   - "capture":   emitted when a piece is captured. The handler receives a *MoveEvent.
//   - "castle":    emitted when a castling move is performed. The handler receives a *MoveEvent.
//   - "enPassant": emitted when an en passant capture occurs. The handler receives a *MoveEvent.
//   - "promote":   emitted when a pawn is promoted. The handler receives the promoted *Square.
//   - "undo":      emitted after a move has been undone. The handler receives the undone *MoveEvent.
//   - "check":     emitted when a player is put in check. The handler receives a *KingThreatEvent.
//   - "checkmate": emitted when a player is checkmated. The handler receives a *KingThreatEvent.
func (c *AlgebraicGameClient) On(ev string, hndlr func(any)) {
	if c == nil {
		return
	}

	c.events.on(ev, hndlr)
}

// Status returns the current status of the game.
// If force is true, it will re-calculate all valid moves and game-end conditions.
func (c *AlgebraicGameClient) Status(frc ...bool) (*GameStatus, error) {
	if len(frc) > 0 && frc[0] {
		if err := c.update(); err != nil {
			return nil, err
		}
	}

	status := &GameStatus{
		Game:         c.game,
		IsCheck:      c.isCheck,
		IsCheckmate:  c.isCheckmate,
		IsRepetition: c.isRepetition,
		IsStalemate:  c.isStalemate,
		NotatedMoves: c.notatedMoves,
	}

	return status, nil
}
