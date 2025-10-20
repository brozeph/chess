package chess

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var notationRegex = regexp.MustCompile(`^[BKQNR]?[a-h]?[1-8]?[x-]?[a-h][1-8][+#]?$`)

type clientStatus struct {
	Board        *Board
	IsCheck      bool
	IsCheckmate  bool
	IsRepetition bool
	IsStalemate  bool
	NotatedMoves map[string]notationMove
}

type notationMove struct {
	Src  *Square
	Dest *Square
}

func getValidMovesByPieceType(pt pieceType, validMoves []validMove) []validMove {
	res := []validMove{}
	for _, mv := range validMoves {
		if mv.Src.Piece != nil && mv.Src.Piece.Type == pt {
			res = append(res, mv)
		}
	}
	return res
}

func getNotationPrefix(src *Square, dest *Square, moves []validMove) string {
	prefix := src.Piece.Notation
	fileCount := map[rune]int{}
	rankCount := map[int]int{}

	for _, mv := range moves {
		for _, sq := range mv.Squares {
			if sq == dest {
				fileCount[mv.Src.File]++
				rankCount[mv.Src.Rank]++
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

func parseNotation(notation string) string {
	if len(notation) < 2 {
		return ""
	}

	dest := notation[len(notation)-2:]
	captureRegex := regexp.MustCompile(`^[a-h]x[a-h][1-8]$`)

	if len(notation) > 2 && captureRegex.MatchString(notation) {
		return dest
	}

	if len(notation) > 2 {
		return notation[:1] + dest
	}

	return ""
}

func sanitizeNotation(n string, usePGN bool) string {
	clean := strings.ReplaceAll(n, "!", "")
	clean = strings.ReplaceAll(clean, "+", "")
	clean = strings.ReplaceAll(clean, "#", "")
	clean = strings.ReplaceAll(clean, "=", "")
	clean = strings.ReplaceAll(clean, "\\", "")
	if usePGN {
		clean = strings.ReplaceAll(clean, "0", "O")
	} else {
		clean = strings.ReplaceAll(clean, "O", "0")
	}
	return clean
}

type AlgebraicClientOptions struct {
	PGN bool
}

type AlgebraicGameClient struct {
	game         *Game
	options      AlgebraicClientOptions
	isCheck      bool
	isCheckmate  bool
	isRepetition bool
	isStalemate  bool
	notatedMoves map[string]notationMove
	validMoves   []validMove
	validation   *gameValidator
	emitter      eventEmitter
}

func CreateAlgebraicGameClient(opts AlgebraicClientOptions) *AlgebraicGameClient {
	g := createGame()
	client := &AlgebraicGameClient{
		game:         g,
		options:      opts,
		notatedMoves: map[string]notationMove{},
		validMoves:   []validMove{},
		validation:   CreateGameValidator(g),
		emitter:      newEventEmitter(),
	}
	client.bindGameEvents()
	client.On("undo", func(interface{}) {
		_ = client.update()
	})
	_ = client.update()
	return client
}

func CreateAlgebraicGameClientFromFEN(fen string, opts AlgebraicClientOptions) (*AlgebraicGameClient, error) {
	if strings.TrimSpace(fen) == "" {
		return nil, errors.New("FEN must be a non-empty string")
	}

	loaded, err := LoadBoard(fen)
	if err != nil {
		return nil, err
	}

	g := createGame()
	g.Board = loaded
	g.Board.LastMovedPiece = nil
	g.hookBoardEvents()

	parts := strings.Split(fen, " ")
	active := "w"
	if len(parts) > 1 {
		active = parts[1]
	}

	baseSide := sideWhite
	if active == "b" {
		baseSide = sideBlack
	}

	whiteFirst := baseSide == sideWhite

	g.sideResolver = func(gm *Game) Side {
		if len(gm.MoveHistory)%2 == 0 {
			if whiteFirst {
				return sideWhite
			}
			return sideBlack
		}
		if whiteFirst {
			return sideBlack
		}
		return sideWhite
	}

	client := &AlgebraicGameClient{
		game:         g,
		options:      opts,
		notatedMoves: map[string]notationMove{},
		validMoves:   []validMove{},
		validation:   CreateGameValidator(g),
		emitter:      newEventEmitter(),
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

func (c *AlgebraicGameClient) emit(event string, data interface{}) {
	c.emitter.emit(event, data)
}

func (c *AlgebraicGameClient) notate(validMoves []validMove) map[string]notationMove {
	algebraic := map[string]notationMove{}

	for _, vm := range validMoves {
		src := vm.Src
		if src.Piece == nil {
			continue
		}
		for _, dest := range vm.Squares {
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
				matches := getValidMovesByPieceType(src.Piece.Type, validMoves)
				if len(matches) > 1 {
					prefix = getNotationPrefix(src, dest, matches)
				} else {
					prefix = src.Piece.Notation
				}
			case pieceKing:
				if src.File == 'e' && dest.File == 'g' {
					if c.options.PGN {
						prefix = "O-O"
					} else {
						prefix = "0-0"
					}
					suffix = ""
				} else if src.File == 'e' && dest.File == 'c' {
					if c.options.PGN {
						prefix = "O-O-O"
					} else {
						prefix = "0-0-0"
					}
					suffix = ""
				} else {
					prefix = src.Piece.Notation
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

			if isPromotion {
				for _, promo := range []string{"R", "N", "B", "Q"} {
					key := prefix + suffix + promo
					algebraic[key] = notationMove{Src: src, Dest: dest}
				}
			} else {
				key := prefix + suffix
				algebraic[key] = notationMove{Src: src, Dest: dest}
			}
		}
	}

	return algebraic
}

func (c *AlgebraicGameClient) On(event string, handler func(interface{})) {
	c.emitter.on(event, handler)
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

func (c *AlgebraicGameClient) CaptureHistory() []*Piece {
	return c.game.CaptureHistory
}

func (c *AlgebraicGameClient) FEN() string {
	return c.game.Board.FEN()
}

func (c *AlgebraicGameClient) Move(notation string, fuzzy bool) (*MoveResult, error) {
	if notation == "" {
		return nil, errors.New("notation is invalid")
	}

	notation = sanitizeNotation(notation, c.options.PGN)

	var promoPiece string
	if len(notation) > 0 {
		last := notation[len(notation)-1]
		if strings.ContainsRune("BNQR", rune(last)) {
			promoPiece = string(last)
		}
	}

	if moveDef, ok := c.notatedMoves[notation]; ok {
		res, err := c.game.move(moveDef.Src, moveDef.Dest, notation)
		if err != nil {
			return nil, err
		}

		if promoPiece != "" {
			var p *Piece
			side := c.game.getCurrentSide().opponent()
			switch promoPiece {
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

	if notationRegex.MatchString(notation) && len(notation) > 1 && !fuzzy {
		return c.Move(parseNotation(notation), true)
	}

	return nil, fmt.Errorf("notation is invalid (%s)", notation)
}

func (c *AlgebraicGameClient) Status(force bool) (*clientStatus, error) {
	if force {
		if err := c.update(); err != nil {
			return nil, err
		}
	}

	status := &clientStatus{
		Board:        c.game.Board,
		IsCheck:      c.isCheck,
		IsCheckmate:  c.isCheckmate,
		IsRepetition: c.isRepetition,
		IsStalemate:  c.isStalemate,
		NotatedMoves: c.notatedMoves,
	}

	return status, nil
}
