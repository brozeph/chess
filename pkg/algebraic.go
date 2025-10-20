package chess

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type validationContext struct {
	destSquares []*Square
	origin      *Square
	p           *Piece
}

type pieceValidator struct {
	board            *Board
	allowBackward    bool
	allowDiagonal    bool
	allowForward     bool
	allowHorizontal  bool
	repeat           int
	pieceType        pieceType
	specialValidator func(v *pieceValidator, ctx *validationContext)
}

func createPieceValidator(pt pieceType, b *Board) *pieceValidator {
	v := &pieceValidator{
		board:     b,
		repeat:    1,
		pieceType: pt,
	}

	switch pt {
	case pieceBishop:
		v.allowDiagonal = true
		v.repeat = 8
	case pieceKing:
		v.allowBackward = true
		v.allowDiagonal = true
		v.allowForward = true
		v.allowHorizontal = true
		v.repeat = 1
	case pieceKnight:
		v.specialValidator = knightSpecial
		v.repeat = 1
	case piecePawn:
		v.allowForward = true
		v.specialValidator = pawnSpecial
		v.repeat = 1
	case pieceQueen:
		v.allowBackward = true
		v.allowDiagonal = true
		v.allowForward = true
		v.allowHorizontal = true
		v.repeat = 8
	case pieceRook:
		v.allowBackward = true
		v.allowForward = true
		v.allowHorizontal = true
		v.repeat = 8
	}

	return v
}

func (v *pieceValidator) start(origin *Square) ([]*Square, error) {
	if origin == nil || origin.Piece == nil || origin.Piece.Type != v.pieceType {
		return nil, errors.New("piece is invalid")
	}

	ctx := &validationContext{
		destSquares: []*Square{},
		origin:      origin,
		p:           origin.Piece,
	}

	findMoveOptions := func(direction neighbor) {
		current := v.board.getNeighborSquare(ctx.origin, direction)
		steps := 0

		for current != nil && steps < v.repeat {
			block := current.Piece != nil && (ctx.p.Type == piecePawn || current.Piece.Side == ctx.p.Side)
			capture := current.Piece != nil && !block

			if !block {
				ctx.destSquares = append(ctx.destSquares, current)
			}

			if capture || block {
				break
			}

			current = v.board.getNeighborSquare(current, direction)
			steps++
		}
	}

	if v.allowForward {
		if ctx.p.Side == sideWhite {
			findMoveOptions(neighborAbove)
		} else {
			findMoveOptions(neighborBelow)
		}
	}

	if v.allowBackward {
		if ctx.p.Side == sideWhite {
			findMoveOptions(neighborBelow)
		} else {
			findMoveOptions(neighborAbove)
		}
	}

	if v.allowHorizontal {
		findMoveOptions(neighborLeft)
		findMoveOptions(neighborRight)
	}

	if v.allowDiagonal {
		findMoveOptions(neighborAboveLeft)
		findMoveOptions(neighborAboveRight)
		findMoveOptions(neighborBelowLeft)
		findMoveOptions(neighborBelowRight)
	}

	if v.specialValidator != nil {
		v.specialValidator(v, ctx)
	}

	return ctx.destSquares, nil
}

func knightSpecial(v *pieceValidator, ctx *validationContext) {
	candidates := []*Square{}

	aboveLeft := v.board.getNeighborSquare(ctx.origin, neighborAboveLeft)
	aboveRight := v.board.getNeighborSquare(ctx.origin, neighborAboveRight)
	belowLeft := v.board.getNeighborSquare(ctx.origin, neighborBelowLeft)
	belowRight := v.board.getNeighborSquare(ctx.origin, neighborBelowRight)

	if aboveLeft != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(aboveLeft, neighborAbove),
			v.board.getNeighborSquare(aboveLeft, neighborLeft),
		)
	}

	if aboveRight != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(aboveRight, neighborAbove),
			v.board.getNeighborSquare(aboveRight, neighborRight),
		)
	}

	if belowLeft != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(belowLeft, neighborBelow),
			v.board.getNeighborSquare(belowLeft, neighborLeft),
		)
	}

	if belowRight != nil {
		candidates = append(candidates,
			v.board.getNeighborSquare(belowRight, neighborBelow),
			v.board.getNeighborSquare(belowRight, neighborRight),
		)
	}

	for _, sq := range candidates {
		if sq == nil {
			continue
		}

		if sq.Piece == nil || sq.Piece.Side != ctx.p.Side {
			ctx.destSquares = append(ctx.destSquares, sq)
		}
	}
}

func pawnSpecial(v *pieceValidator, ctx *validationContext) {
	var diagLeft, diagRight *Square
	if ctx.p.Side == sideWhite {
		diagLeft = v.board.getNeighborSquare(ctx.origin, neighborAboveLeft)
		diagRight = v.board.getNeighborSquare(ctx.origin, neighborAboveRight)
	} else {
		diagLeft = v.board.getNeighborSquare(ctx.origin, neighborBelowLeft)
		diagRight = v.board.getNeighborSquare(ctx.origin, neighborBelowRight)
	}

	for _, sq := range []*Square{diagLeft, diagRight} {
		if sq == nil {
			continue
		}
		if sq.Piece != nil && sq.Piece.Side != ctx.p.Side {
			ctx.destSquares = append(ctx.destSquares, sq)
		}
	}

	if ctx.p.MoveCount == 0 && len(ctx.destSquares) > 0 {
		firstForward := ctx.destSquares[0]
		if firstForward != nil && firstForward.Piece == nil {
			var further *Square
			if ctx.p.Side == sideWhite {
				further = v.board.getNeighborSquare(firstForward, neighborAbove)
			} else {
				further = v.board.getNeighborSquare(firstForward, neighborBelow)
			}

			if further != nil && further.Piece == nil {
				ctx.destSquares = append(ctx.destSquares, further)
			}
		}
	}

	rankForEnPassant := 5
	if ctx.p.Side == sideBlack {
		rankForEnPassant = 4
	}

	if ctx.origin.Rank != rankForEnPassant {
		return
	}

	left := v.board.getNeighborSquare(ctx.origin, neighborLeft)
	right := v.board.getNeighborSquare(ctx.origin, neighborRight)

	for _, adj := range []*Square{left, right} {
		if adj == nil || adj.Piece == nil {
			continue
		}

		if adj.Piece.Type == piecePawn &&
			adj.Piece.Side != ctx.p.Side &&
			adj.Piece.MoveCount == 1 &&
			v.board.LastMovedPiece == adj.Piece {
			var captureTarget *Square
			if adj.Piece.Side == sideBlack {
				captureTarget = v.board.getNeighborSquare(adj, neighborAbove)
			} else {
				captureTarget = v.board.getNeighborSquare(adj, neighborBelow)
			}
			if captureTarget != nil {
				ctx.destSquares = append(ctx.destSquares, captureTarget)
			}
		}
	}
}

type validMove struct {
	Src     *Square
	Squares []*Square
}

type boardValidation struct {
	game  *Game
	board *Board
}

func createBoardValidation(g *Game) *boardValidation {
	return &boardValidation{
		game:  g,
		board: g.Board,
	}
}

func (v *boardValidation) evaluateCastle(validMoves []validMove) {
	getValidSquares := func(src *Square) []*Square {
		for _, vm := range validMoves {
			if vm.Src == src {
				return vm.Squares
			}
		}
		return nil
	}

	rank := 1
	if v.game.getCurrentSide() == sideBlack {
		rank = 8
	}

	squares := map[rune]*Square{
		'a': v.board.getSquare('a', rank),
		'b': v.board.getSquare('b', rank),
		'c': v.board.getSquare('c', rank),
		'd': v.board.getSquare('d', rank),
		'e': v.board.getSquare('e', rank),
		'f': v.board.getSquare('f', rank),
		'g': v.board.getSquare('g', rank),
		'h': v.board.getSquare('h', rank),
	}

	kingSquare := squares['e']
	if kingSquare == nil || kingSquare.Piece == nil {
		return
	}
	if kingSquare.Piece.Type != pieceKing || kingSquare.Piece.MoveCount != 0 {
		return
	}
	if v.isSquareAttacked(kingSquare) {
		return
	}

	// queen side
	if squares['a'] != nil &&
		squares['a'].Piece != nil &&
		squares['a'].Piece.Type == pieceRook &&
		squares['a'].Piece.MoveCount == 0 {

		if squares['b'].Piece == nil &&
			squares['c'].Piece == nil &&
			squares['d'].Piece == nil {

			res1, _ := v.board.Move(squares['e'], squares['d'], true, "")
			if !v.isSquareAttacked(squares['d']) {
				res1.Undo()
				res2, _ := v.board.Move(squares['e'], squares['c'], true, "")
				if !v.isSquareAttacked(squares['c']) {
					moveSquares := getValidSquares(squares['e'])
					if moveSquares != nil {
						moveSquares = append(moveSquares, squares['c'])
						for idx := range validMoves {
							if validMoves[idx].Src == squares['e'] {
								validMoves[idx].Squares = moveSquares
							}
						}
					}
				}
				res2.Undo()
			} else {
				res1.Undo()
			}
		}
	}

	// king side
	if squares['h'] != nil &&
		squares['h'].Piece != nil &&
		squares['h'].Piece.Type == pieceRook &&
		squares['h'].Piece.MoveCount == 0 {

		if squares['f'].Piece == nil &&
			squares['g'].Piece == nil {

			res1, _ := v.board.Move(squares['e'], squares['f'], true, "")
			if !v.isSquareAttacked(squares['f']) {
				res1.Undo()
				res2, _ := v.board.Move(squares['e'], squares['g'], true, "")
				if !v.isSquareAttacked(squares['g']) {
					moveSquares := getValidSquares(squares['e'])
					if moveSquares != nil {
						moveSquares = append(moveSquares, squares['g'])
						for idx := range validMoves {
							if validMoves[idx].Src == squares['e'] {
								validMoves[idx].Squares = moveSquares
							}
						}
					}
				}
				res2.Undo()
			} else {
				res1.Undo()
			}
		}
	}
}

func (v *boardValidation) filterKingAttack(kingSquare *Square, moves []validMove) []validMove {
	filtered := make([]validMove, 0, len(moves))

	for _, mv := range moves {
		validSquares := []*Square{}
		for _, dest := range mv.Squares {
			res, err := v.board.Move(mv.Src, dest, true, "")
			if err != nil {
				continue
			}

			isCheck := false
			if res.Move.Piece.Type != pieceKing {
				isCheck = v.isSquareAttacked(kingSquare)
			} else {
				isCheck = v.isSquareAttacked(dest)
			}

			res.Undo()

			if !isCheck {
				validSquares = append(validSquares, dest)
			}
		}

		if len(validSquares) > 0 {
			filtered = append(filtered, validMove{
				Src:     mv.Src,
				Squares: validSquares,
			})
		}
	}

	return filtered
}

type attackContext struct {
	attacked bool
	piece    *Piece
	square   *Square
}

type kingThreatEvent struct {
	AttackingSquare *Square
	KingSquare      *Square
}

func (v *boardValidation) findAttackers(target *Square) []attackContext {
	if target == nil || target.Piece == nil {
		return []attackContext{}
	}

	results := []attackContext{}

	checkDirection := func(nb neighbor) {
		current := v.board.getNeighborSquare(target, nb)
		for current != nil {
			if current.Piece != nil {
				if current.Piece.Side != target.Piece.Side {
					validator := createPieceValidator(current.Piece.Type, v.board)
					destSquares, err := validator.start(current)
					if err == nil {
						for _, sq := range destSquares {
							if sq == target {
								results = append(results, attackContext{
									attacked: true,
									piece:    current.Piece,
									square:   current,
								})
								return
							}
						}
					}
				}
				return
			}
			current = v.board.getNeighborSquare(current, nb)
		}
	}

	checkKnight := func(nb neighbor) {
		current := v.board.getNeighborSquare(target, nb)
		if current == nil || current.Piece == nil {
			return
		}
		if current.Piece.Type != pieceKnight || current.Piece.Side == target.Piece.Side {
			return
		}
		validator := createPieceValidator(pieceKnight, v.board)
		destSquares, err := validator.start(current)
		if err != nil {
			return
		}
		for _, sq := range destSquares {
			if sq == target {
				results = append(results, attackContext{
					attacked: true,
					piece:    current.Piece,
					square:   current,
				})
				return
			}
		}
	}

	directions := []neighbor{
		neighborAbove,
		neighborAboveRight,
		neighborRight,
		neighborBelowRight,
		neighborBelow,
		neighborBelowLeft,
		neighborLeft,
		neighborAboveLeft,
	}

	for _, dir := range directions {
		checkDirection(dir)
	}

	knightDirections := []neighbor{
		neighborKnightAboveRight,
		neighborKnightRightAbove,
		neighborKnightBelowRight,
		neighborKnightRightBelow,
		neighborKnightBelowLeft,
		neighborKnightLeftBelow,
		neighborKnightAboveLeft,
		neighborKnightLeftAbove,
	}

	for _, dir := range knightDirections {
		checkKnight(dir)
	}

	return results
}

func (v *boardValidation) isSquareAttacked(sq *Square) bool {
	return len(v.findAttackers(sq)) > 0
}

func (v *boardValidation) start() ([]validMove, error) {
	if v.board == nil {
		return nil, errors.New("board is invalid")
	}

	squares := v.board.getSquares(v.game.getCurrentSide())
	validMoves := []validMove{}
	var kingSquare *Square

	for _, sq := range squares {
		if sq.Piece == nil {
			continue
		}

		if sq.Piece.Type == pieceKing {
			kingSquare = sq
		}

		validator := createPieceValidator(sq.Piece.Type, v.board)
		destSquares, err := validator.start(sq)
		if err != nil {
			return nil, err
		}

		if len(destSquares) > 0 {
			validMoves = append(validMoves, validMove{
				Src:     sq,
				Squares: destSquares,
			})
		}
	}

	v.evaluateCastle(validMoves)
	validMoves = v.filterKingAttack(kingSquare, validMoves)

	attackers := v.findAttackers(kingSquare)
	for _, attacker := range attackers {
		data := kingThreatEvent{
			AttackingSquare: attacker.square,
			KingSquare:      kingSquare,
		}
		if len(validMoves) == 0 {
			v.game.emit("checkmate", data)
		} else {
			v.game.emit("check", data)
		}
	}

	return validMoves, nil
}

type gameValidation struct {
	game *Game
}

func createGameValidation(g *Game) *gameValidation {
	return &gameValidation{game: g}
}

func (gv *gameValidation) findKingSquare(sd Side) *Square {
	squares := gv.game.Board.getSquares(sd)
	for _, sq := range squares {
		if sq.Piece != nil && sq.Piece.Type == pieceKing {
			return sq
		}
	}
	return nil
}

func (gv *gameValidation) isRepetition() bool {
	counts := map[string]int{}
	for _, mv := range gv.game.MoveHistory {
		counts[mv.hashCode]++
		if counts[mv.hashCode] == 3 {
			return true
		}
	}
	return false
}

type validationResult struct {
	IsCheck      bool
	IsCheckmate  bool
	IsRepetition bool
	IsStalemate  bool
	ValidMoves   []validMove
}

func (gv *gameValidation) start() (*validationResult, error) {
	result := &validationResult{
		IsCheck:      false,
		IsCheckmate:  false,
		IsRepetition: false,
		IsStalemate:  false,
		ValidMoves:   []validMove{},
	}

	bv := createBoardValidation(gv.game)
	validMoves, err := bv.start()
	if err != nil {
		return nil, err
	}

	kingSquare := gv.findKingSquare(gv.game.getCurrentSide())
	isAttacked := bv.isSquareAttacked(kingSquare)

	result.IsCheck = isAttacked && len(validMoves) > 0
	result.IsCheckmate = isAttacked && len(validMoves) == 0
	result.IsStalemate = !isAttacked && len(validMoves) == 0
	result.ValidMoves = validMoves
	result.IsRepetition = gv.isRepetition()

	return result, nil
}

type notationMove struct {
	Src  *Square
	Dest *Square
}

type algebraicClientOptions struct {
	PGN bool
}

type algebraicGameClient struct {
	game         *Game
	options      algebraicClientOptions
	isCheck      bool
	isCheckmate  bool
	isRepetition bool
	isStalemate  bool
	notatedMoves map[string]notationMove
	validMoves   []validMove
	validation   *gameValidation
	emitter      eventEmitter
}

func createAlgebraicGameClient(opts algebraicClientOptions) *algebraicGameClient {
	g := createGame()
	client := &algebraicGameClient{
		game:         g,
		options:      opts,
		notatedMoves: map[string]notationMove{},
		validMoves:   []validMove{},
		validation:   createGameValidation(g),
		emitter:      newEventEmitter(),
	}
	client.bindGameEvents()
	client.On("undo", func(interface{}) {
		_ = client.update()
	})
	_ = client.update()
	return client
}

func (c *algebraicGameClient) bindGameEvents() {
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

func (c *algebraicGameClient) On(event string, handler func(interface{})) {
	c.emitter.on(event, handler)
}

func (c *algebraicGameClient) emit(event string, data interface{}) {
	c.emitter.emit(event, data)
}

func createAlgebraicGameClientFromFEN(fen string, opts algebraicClientOptions) (*algebraicGameClient, error) {
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

	client := &algebraicGameClient{
		game:         g,
		options:      opts,
		notatedMoves: map[string]notationMove{},
		validMoves:   []validMove{},
		validation:   createGameValidation(g),
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

func (c *algebraicGameClient) update() error {
	result, err := c.validation.start()
	if err != nil {
		return err
	}
	c.isCheck = result.IsCheck
	c.isCheckmate = result.IsCheckmate
	c.isRepetition = result.IsRepetition
	c.isStalemate = result.IsStalemate
	c.validMoves = result.ValidMoves
	c.notatedMoves = notate(result.ValidMoves, c)
	return nil
}

type clientStatus struct {
	Board        *Board
	IsCheck      bool
	IsCheckmate  bool
	IsRepetition bool
	IsStalemate  bool
	NotatedMoves map[string]notationMove
}

func (c *algebraicGameClient) getStatus(force bool) (*clientStatus, error) {
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

func (c *algebraicGameClient) getFEN() string {
	return c.game.Board.FEN()
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

var notationRegex = regexp.MustCompile(`^[BKQNR]?[a-h]?[1-8]?[x-]?[a-h][1-8][+#]?$`)

func (c *algebraicGameClient) move(notation string, fuzzy bool) (*MoveResult, error) {
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
		return c.move(parseNotation(notation), true)
	}

	return nil, fmt.Errorf("notation is invalid (%s)", notation)
}

func (c *algebraicGameClient) getCaptureHistory() []*Piece {
	return c.game.CaptureHistory
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

func notate(validMoves []validMove, client *algebraicGameClient) map[string]notationMove {
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
					if client.options.PGN {
						prefix = "O-O"
					} else {
						prefix = "0-0"
					}
					suffix = ""
				} else if src.File == 'e' && dest.File == 'c' {
					if client.options.PGN {
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

func pieceSymbol(p *Piece) rune {
	if p == nil {
		return '.'
	}

	var symbol rune
	switch p.Type {
	case piecePawn:
		symbol = 'p'
	case pieceKnight:
		symbol = 'n'
	case pieceBishop:
		symbol = 'b'
	case pieceRook:
		symbol = 'r'
	case pieceQueen:
		symbol = 'q'
	case pieceKing:
		symbol = 'k'
	default:
		symbol = '?'
	}

	if p.Side == sideWhite {
		return rune(strings.ToUpper(string(symbol))[0])
	}

	return symbol
}

func printBoard(b *Board) {
	fmt.Println("Current board position:")
	for rank := 8; rank >= 1; rank-- {
		fmt.Printf("%d ", rank)
		for file := 'a'; file <= 'h'; file++ {
			sq := b.getSquare(file, rank)
			fmt.Printf("%c ", pieceSymbol(sq.Piece))
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

	opts := algebraicClientOptions{PGN: usePGN}

	var client *algebraicGameClient
	var err error

	if strings.TrimSpace(fen) != "" {
		client, err = createAlgebraicGameClientFromFEN(fen, opts)
		if err != nil {
			log.Fatalf("failed to load FEN: %v", err)
		}
	} else {
		client = createAlgebraicGameClient(opts)
	}

	if movesArg != "" {
		for _, mv := range strings.Fields(movesArg) {
			if _, err := client.move(mv, false); err != nil {
				log.Fatalf("failed to apply move %s: %v", mv, err)
			}
		}
	}

	status, err := client.getStatus(false)
	if err != nil {
		log.Fatalf("failed to evaluate position: %v", err)
	}

	printBoard(status.Board)
	fmt.Printf("FEN: %s\n", client.getFEN())
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
