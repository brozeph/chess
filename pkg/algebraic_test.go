package chess

import "testing"

func mustMove(t *testing.T, client *algebraicGameClient, notation string) *MoveResult {
	t.Helper()
	res, err := client.move(notation, false)
	if err != nil {
		t.Fatalf("move %s failed: %v", notation, err)
	}
	return res
}

func mustStatus(t *testing.T, client *algebraicGameClient, force bool) *clientStatus {
	t.Helper()
	sts, err := client.getStatus(force)
	if err != nil {
		t.Fatalf("getStatus failed: %v", err)
	}
	return sts
}

func TestCreateInitialStatus(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	status := mustStatus(t, client, false)

	if status.IsCheck {
		t.Fatalf("expected IsCheck false")
	}
	if status.IsCheckmate {
		t.Fatalf("expected IsCheckmate false")
	}
	if status.IsRepetition {
		t.Fatalf("expected IsRepetition false")
	}
	if status.IsStalemate {
		t.Fatalf("expected IsStalemate false")
	}
	if got := len(status.NotatedMoves); got != 20 {
		t.Fatalf("expected 20 notated moves, got %d", got)
	}
}

func TestMoveEventTriggered(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	var events []*Move

	client.On("move", func(data interface{}) {
		if mv, ok := data.(*Move); ok {
			events = append(events, mv)
		}
	})

	mustMove(t, client, "b4")
	mustMove(t, client, "e6")

	if len(events) != 2 {
		t.Fatalf("expected 2 move events, got %d", len(events))
	}
}

func TestStatusAfterMoves(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "b4")
	mustMove(t, client, "e6")

	status := mustStatus(t, client, false)

	if status.IsCheck || status.IsCheckmate || status.IsRepetition || status.IsStalemate {
		t.Fatalf("unexpected flags %#v", status)
	}
	if got := len(status.NotatedMoves); got != 21 {
		t.Fatalf("expected 21 notated moves, got %d", got)
	}
}

func TestPawnCapture(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	result := mustMove(t, client, "exd5")

	if result.Move.CapturedPiece == nil || result.Move.CapturedPiece.Type != piecePawn {
		t.Fatalf("expected captured pawn, got %+v", result.Move.CapturedPiece)
	}
}

func TestCaptureEvent(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	var events []*Move

	client.On("capture", func(data interface{}) {
		if mv, ok := data.(*Move); ok {
			events = append(events, mv)
		}
	})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	mustMove(t, client, "exd5")

	if len(events) != 1 {
		t.Fatalf("expected 1 capture event, got %d", len(events))
	}
}

func TestMoveHistoryRecordsNotation(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	mustMove(t, client, "exd5")

	if len(client.game.MoveHistory) < 3 {
		t.Fatalf("expected at least 3 moves, got %d", len(client.game.MoveHistory))
	}
	if client.game.MoveHistory[2].Algebraic != "exd5" {
		t.Fatalf("expected notation exd5, got %s", client.game.MoveHistory[2].Algebraic)
	}
}

func TestCaptureHistoryAndUndo(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	capture := mustMove(t, client, "exd5")

	history := client.getCaptureHistory()
	if len(history) != 1 {
		t.Fatalf("expected capture history length 1, got %d", len(history))
	}
	if history[0].Type != piecePawn {
		t.Fatalf("expected pawn capture history, got %v", history[0].Type)
	}

	capture.Undo()

	history = client.getCaptureHistory()
	if len(history) != 0 {
		t.Fatalf("expected empty capture history after undo, got %d", len(history))
	}
}

func TestKnightDisambiguation(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "Nc3")
	mustMove(t, client, "Nf6")
	mustMove(t, client, "Nd5")
	mustMove(t, client, "Ng8")
	mustMove(t, client, "Nf4")
	mustMove(t, client, "Nf6")

	status := mustStatus(t, client, false)

	if _, ok := status.NotatedMoves["Nfh3"]; !ok {
		t.Fatalf("expected Nfh3 in notated moves")
	}
	if _, ok := status.NotatedMoves["Ngh3"]; !ok {
		t.Fatalf("expected Ngh3 in notated moves")
	}
}

func TestRookDisambiguationRanks(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "a4")
	mustMove(t, client, "a5")
	mustMove(t, client, "h4")
	mustMove(t, client, "h5")
	mustMove(t, client, "Ra3")
	mustMove(t, client, "Ra6")
	mustMove(t, client, "Rhh3")
	mustMove(t, client, "Rhh6")

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["Rae3"]; !ok {
		t.Fatalf("expected Rae3")
	}
	if _, ok := status.NotatedMoves["Rhe3"]; !ok {
		t.Fatalf("expected Rhe3")
	}
}

func TestRookDisambiguationFiles(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	sequence := []string{
		"a4", "a5", "h4", "h5", "Ra3", "Ra6", "Rhh3", "Rhh6",
		"Rae3", "Rh8", "Re6", "Ra8", "Rhe3", "Ra6",
	}

	for _, mv := range sequence {
		mustMove(t, client, mv)
	}

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["R6e5"]; !ok {
		t.Fatalf("expected R6e5")
	}
	if _, ok := status.NotatedMoves["R3e5"]; !ok {
		t.Fatalf("expected R3e5")
	}
}

func TestWhiteCastleLeftEvent(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	var events []*Move

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*Move); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.getSquare('b', 1).Piece = nil
	client.game.Board.getSquare('c', 1).Piece = nil
	client.game.Board.getSquare('d', 1).Piece = nil

	status := mustStatus(t, client, true)
	if _, ok := status.NotatedMoves["0-0-0"]; !ok {
		t.Fatalf("expected 0-0-0 notation")
	}

	mustMove(t, client, "0-0-0")

	if len(events) != 1 {
		t.Fatalf("expected castle event")
	}
}

func TestWhiteCastleLeftPGN(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{PGN: true})
	var events []*Move

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*Move); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.getSquare('b', 1).Piece = nil
	client.game.Board.getSquare('c', 1).Piece = nil
	client.game.Board.getSquare('d', 1).Piece = nil

	status := mustStatus(t, client, true)
	if _, ok := status.NotatedMoves["O-O-O"]; !ok {
		t.Fatalf("expected O-O-O notation")
	}

	mustMove(t, client, "O-O-O")

	if len(events) != 1 {
		t.Fatalf("expected castle event")
	}
}

func TestBlackCastleRightEvent(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	var events []*Move

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*Move); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.getSquare('f', 8).Piece = nil
	client.game.Board.getSquare('g', 8).Piece = nil
	mustStatus(t, client, true)
	mustMove(t, client, "a4")
	status := mustStatus(t, client, false)

	if _, ok := status.NotatedMoves["0-0"]; !ok {
		t.Fatalf("expected 0-0 notation")
	}

	mustMove(t, client, "0-0")

	if len(events) != 1 {
		t.Fatalf("expected castle event")
	}
}

func TestBlackCastleRightPGN(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{PGN: true})
	var events []*Move

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*Move); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.getSquare('f', 8).Piece = nil
	client.game.Board.getSquare('g', 8).Piece = nil
	mustStatus(t, client, true)
	mustMove(t, client, "a4")
	status := mustStatus(t, client, false)

	if _, ok := status.NotatedMoves["O-O"]; !ok {
		t.Fatalf("expected O-O notation")
	}

	mustMove(t, client, "O-O")

	if len(events) != 1 {
		t.Fatalf("expected castle event")
	}
}

func TestParseWhiteCastleLeft(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	client.game.Board.getSquare('b', 1).Piece = nil
	client.game.Board.getSquare('c', 1).Piece = nil
	client.game.Board.getSquare('d', 1).Piece = nil
	mustStatus(t, client, true)

	res := mustMove(t, client, "O-O-O")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}

func TestParseWhiteCastleLeftPGN(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{PGN: true})

	client.game.Board.getSquare('b', 1).Piece = nil
	client.game.Board.getSquare('c', 1).Piece = nil
	client.game.Board.getSquare('d', 1).Piece = nil
	mustStatus(t, client, true)

	res := mustMove(t, client, "0-0-0")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}

func TestParseBlackCastleRight(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	client.game.Board.getSquare('f', 8).Piece = nil
	client.game.Board.getSquare('g', 8).Piece = nil
	mustStatus(t, client, true)
	mustMove(t, client, "a4")

	res := mustMove(t, client, "O-O")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}

func TestParseBlackCastleRightPGN(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	client.game.Board.getSquare('f', 8).Piece = nil
	client.game.Board.getSquare('g', 8).Piece = nil
	mustStatus(t, client, true)
	mustMove(t, client, "a4")

	res := mustMove(t, client, "0-0")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}

func TestWhitePawnPromotionMoves(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	client.game.Board.getSquare('a', 7).Piece = nil
	client.game.Board.getSquare('a', 8).Piece = nil
	client.game.Board.getSquare('a', 2).Piece = nil
	client.game.Board.getSquare('a', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.getSquare('a', 7).Piece.MoveCount = 1

	status := mustStatus(t, client, true)

	if _, ok := status.NotatedMoves["a8"]; ok {
		t.Fatalf("expected base move to require promotion")
	}

	for _, move := range []string{"a8R", "a8N", "a8B", "a8Q"} {
		if _, ok := status.NotatedMoves[move]; !ok {
			t.Fatalf("expected promotion move %s", move)
		}
	}
}

func TestBlackPawnPromotionMoves(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	client.game.Board.getSquare('a', 2).Piece = nil
	client.game.Board.getSquare('a', 1).Piece = nil
	client.game.Board.getSquare('a', 7).Piece = nil
	client.game.Board.getSquare('a', 2).Piece = newPiece(piecePawn, sideBlack)
	client.game.Board.getSquare('a', 2).Piece.MoveCount = 1

	mustStatus(t, client, true)
	mustMove(t, client, "h4")
	status := mustStatus(t, client, true)

	if _, ok := status.NotatedMoves["a1"]; ok {
		t.Fatalf("expected base move to require promotion")
	}

	for _, move := range []string{"a1R", "a1N", "a1B", "a1Q"} {
		if _, ok := status.NotatedMoves[move]; !ok {
			t.Fatalf("expected promotion move %s", move)
		}
	}
}

func TestWhitePawnPromotionExecution(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	for _, sq := range []string{"a7", "a8", "b8", "c8", "d8", "a2"} {
		client.game.Board.getSquareByName(sq).Piece = nil
	}
	client.game.Board.getSquare('a', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.getSquare('a', 7).Piece.MoveCount = 1

	mustStatus(t, client, true)
	res := mustMove(t, client, "a8R")
	status := mustStatus(t, client, false)

	if res.Move.PostSquare.Piece == nil || res.Move.PostSquare.Piece.Type != pieceRook {
		t.Fatalf("expected rook on promotion square")
	}
	if !status.IsCheckmate {
		t.Fatalf("expected checkmate true")
	}
	if len(client.game.MoveHistory) == 0 || !client.game.MoveHistory[0].Promotion {
		t.Fatalf("expected promotion flag")
	}
}

func TestBlackPawnPromotionExecution(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	for _, sq := range []string{"a2", "a1", "b1", "c1", "d1", "a7"} {
		client.game.Board.getSquareByName(sq).Piece = nil
	}
	client.game.Board.getSquare('a', 2).Piece = newPiece(piecePawn, sideBlack)
	client.game.Board.getSquare('a', 2).Piece.MoveCount = 1

	mustStatus(t, client, true)
	mustMove(t, client, "h3")
	res := mustMove(t, client, "a1R")
	status := mustStatus(t, client, false)

	if res.Move.PostSquare.Piece == nil || res.Move.PostSquare.Piece.Type != pieceRook {
		t.Fatalf("expected rook on promotion square")
	}
	if !status.IsCheckmate {
		t.Fatalf("expected checkmate true")
	}
	if len(client.game.MoveHistory) < 2 || client.game.MoveHistory[0].Promotion || !client.game.MoveHistory[1].Promotion {
		t.Fatalf("expected promotion flagged on second move")
	}
}

func TestPromotionEvent(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	var events []*Square

	client.On("promote", func(data interface{}) {
		if sq, ok := data.(*Square); ok {
			events = append(events, sq)
		}
	})

	for _, sq := range []string{"a7", "a8", "b8", "c8", "d8", "a2"} {
		client.game.Board.getSquareByName(sq).Piece = nil
	}
	client.game.Board.getSquare('a', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.getSquare('a', 7).Piece.MoveCount = 1

	mustStatus(t, client, true)
	mustMove(t, client, "a8R")

	if len(client.game.MoveHistory) == 0 || !client.game.MoveHistory[0].Promotion {
		t.Fatalf("expected promotion flag")
	}
	if len(events) != 1 {
		t.Fatalf("expected promotion event")
	}
}

func TestAmbiguousNotationThrows(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	sequence := []string{
		"a4", "a5", "h4", "h5", "Ra3", "Ra6",
	}
	for _, mv := range sequence {
		mustMove(t, client, mv)
	}

	if _, err := client.move("Rh3", false); err == nil {
		t.Fatalf("expected ambiguous notation error")
	}
}

func TestInvalidNotationThrows(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	if _, err := client.move("h6", false); err == nil {
		t.Fatalf("expected move error for h6")
	}
	if _, err := client.move("z9", false); err == nil {
		t.Fatalf("expected move error for z9")
	}
}

func TestVerboseNotationParses(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	res := mustMove(t, client, "Nb1c3")
	if res.Move.PostSquare.File != 'c' || res.Move.PostSquare.Rank != 3 {
		t.Fatalf("expected knight on c3")
	}
	if res.Move.PostSquare.Piece == nil || res.Move.PostSquare.Piece.Type != pieceKnight {
		t.Fatalf("expected knight piece")
	}
}

func TestIssue1NoPhantomPawn(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	target := client.game.Board.getSquare('c', 5)

	moves := []string{
		"e4", "e5", "Nf3", "Nc6", "Bb5", "Nf6",
		"O-O", "Nxe4", "d4", "Nd6",
	}
	for _, mv := range moves {
		mustMove(t, client, mv)
	}

	if target.Piece != nil {
		t.Fatalf("expected no piece on c5 before Bxc6")
	}

	mustMove(t, client, "Bxc6")
	if target.Piece != nil {
		t.Fatalf("expected no piece on c5 after Bxc6")
	}
}

func TestIssue3NoPhantomPawn(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	target := client.game.Board.getSquare('a', 6)

	moves := []string{
		"e4", "e5", "d3", "Nc6", "Nf3", "Bb4",
		"Nfd2", "d6", "a3", "Bc5", "Be2", "Qf6",
		"0-0", "Bxf2", "Rxf2", "Qe6", "Nc4", "Nd4",
		"Bf1", "Bd7", "c3", "Nb3", "Ra2", "Ba4",
		"Qc2", "Nh6", "d4", "Ng4", "Rf3", "b5",
		"Nxe5", "Nxc1", "Qxc1", "dxe5", "Ra1", "Rb8",
		"h3", "Rb6", "hxg4", "Qxg4", "Nd2", "a5",
		"dxe5", "Rc6", "c4", "h5", "Rb1", "Rhh6",
		"Ra1", "Rce6", "Bd3", "Rxe5", "cxb5",
	}

	for _, mv := range moves {
		mustMove(t, client, mv)
	}

	if target.Piece != nil {
		t.Fatalf("expected no piece on a6 before Rg6")
	}

	mustMove(t, client, "Rg6")

	if target.Piece != nil {
		t.Fatalf("expected no piece on a6 after Rg6")
	}
}

func TestIssue4CheckmateDetection(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	moves := []string{
		"e4", "e5", "Nc3", "d6", "Bc4", "Be6",
		"Bb3", "Nf6", "Nge2", "Nh5", "Bxe6", "fxe6",
		"d4", "Be7", "dxe5", "dxe5", "Qxd8", "Bxd8",
		"Be3", "0-0", "0-0-0", "Nc6", "Rhf1", "Bh4",
		"Nb5", "Rac8", "f3", "a6", "Nbc3", "Nb4",
		"Bc5", "Nxa2", "Nxa2", "b6", "Bxf8", "Rxf8",
		"Nb4", "a5", "Nc6", "Ra8", "Nxe5", "c5",
		"Rd6", "Rc8", "Rxb6", "c4", "f4", "c3",
		"Nxc3", "Rxc3", "Rb8",
	}

	for _, mv := range moves {
		mustMove(t, client, mv)
	}

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["Kf7"]; ok {
		t.Fatalf("expected Kf7 not to be available")
	}
}

func TestIssue8NoPhantomPawn(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	target := client.game.Board.getSquare('e', 6)

	mustMove(t, client, "d4")
	mustMove(t, client, "a6")
	mustMove(t, client, "d5")

	if target.Piece != nil {
		t.Fatalf("expected no piece on e6 before e5")
	}

	mustMove(t, client, "e5")

	if target.Piece != nil {
		t.Fatalf("expected no piece on e6 after e5")
	}
}

func TestIssue15PawnAdvance(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "a5")
	mustMove(t, client, "Ba6")

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["b5"]; !ok {
		t.Fatalf("expected pawn to advance two squares")
	}
}

func TestIssue17PromotionAvailability(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	client.game.Board.getSquare('c', 7).Piece = nil
	client.game.Board.getSquare('c', 8).Piece = nil
	client.game.Board.getSquare('c', 2).Piece = nil
	client.game.Board.getSquare('c', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.getSquare('c', 7).Piece.MoveCount = 1
	client.game.Board.getSquare('h', 7).Piece = nil
	client.game.Board.getSquare('h', 7).Piece = newPiece(pieceBishop, sideWhite)
	client.game.Board.getSquare('h', 7).Piece.MoveCount = 1

	status := mustStatus(t, client, true)

	for _, mv := range []string{"cxb8R", "cxb8N", "cxb8B", "cxb8Q", "cxd8R", "cxd8N", "cxd8B", "cxd8Q"} {
		if _, ok := status.NotatedMoves[mv]; !ok {
			t.Fatalf("expected promotion move %s", mv)
		}
	}
	if _, ok := status.NotatedMoves["Bxg8R"]; ok {
		t.Fatalf("bishop should not have promotion move")
	}
}

func TestIssue18PromotionAvailability(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	client.game.Board.getSquare('c', 7).Piece = nil
	client.game.Board.getSquare('c', 2).Piece = nil
	client.game.Board.getSquare('c', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.getSquare('c', 7).Piece.MoveCount = 1

	status := mustStatus(t, client, true)

	if _, ok := status.NotatedMoves["cxb8"]; ok {
		t.Fatalf("expected base move to require promotion")
	}
	for _, mv := range []string{"cxb8Q", "cxb8R", "cxb8B", "cxb8N"} {
		if _, ok := status.NotatedMoves[mv]; !ok {
			t.Fatalf("expected promotion %s", mv)
		}
	}
}

func TestIssue23CheckEvent(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	var event kingThreatEvent
	triggered := false

	client.On("check", func(data interface{}) {
		if ev, ok := data.(kingThreatEvent); ok {
			event = ev
			triggered = true
		}
	})

	client.game.Board.getSquare('b', 1).Piece = nil
	client.game.Board.getSquare('f', 6).Piece = newPiece(pieceKnight, sideWhite)
	client.game.Board.getSquare('f', 6).Piece.MoveCount = 1

	mustMove(t, client, "a3")
	status := mustStatus(t, client, true)

	if !triggered {
		t.Fatalf("expected check event")
	}
	if event.AttackingSquare == nil || event.AttackingSquare.Piece == nil || event.AttackingSquare.Piece.Type != pieceKnight {
		t.Fatalf("expected attacking knight")
	}
	if _, ok := status.NotatedMoves["exf6"]; !ok {
		t.Fatalf("expected exf6 capture")
	}
	if _, ok := status.NotatedMoves["gxf6"]; !ok {
		t.Fatalf("expected gxf6 capture")
	}
	if _, ok := status.NotatedMoves["Nxf6"]; !ok {
		t.Fatalf("expected Nxf6 capture")
	}
}

func TestIssue43ParseGxf3Check(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	moves := []string{
		"d4", "d6", "e4", "Nf6", "Nc3", "e5", "Nf3", "Nbd7",
		"Bc4", "Nb6", "dxe5", "Nxc4", "exf6", "Qxf6", "Bg5", "Nxb2",
		"Qd2", "Qe6", "Nd5", "Qxe4+", "Kf1", "Qc4+", "Kg1", "Be6",
		"Ne3", "Qc5", "Rb1", "Na4", "c4", "Nb6", "Qb2", "h6",
		"Bh4", "Rg8", "Nd4", "g5", "Nxe6", "fxe6", "Qf6", "Qe5",
		"Qxe5", "dxe5", "Bg3", "O-O-O", "Bxe5", "Bc5", "h4", "g4",
		"Kh2", "Rd2", "Kg3", "Nd7", "Bb2", "Bd6+", "f4", "gxf3+",
		"Kxf3", "Rg3+", "Ke4", "Nc5+",
	}

	for _, mv := range moves {
		mustMove(t, client, mv)
	}

	status := mustStatus(t, client, false)
	if !status.IsCheckmate {
		t.Fatalf("expected checkmate true")
	}
}

func TestIssue53EnPassantEvent(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})
	var events []*Move

	client.On("enPassant", func(data interface{}) {
		if mv, ok := data.(*Move); ok {
			events = append(events, mv)
		}
	})

	mustMove(t, client, "e4")
	mustMove(t, client, "d5")
	mustMove(t, client, "e5")
	mustMove(t, client, "f5")

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["f6"]; ok {
		t.Fatalf("unexpected f6 move")
	}
	if _, ok := status.NotatedMoves["exf6"]; !ok {
		t.Fatalf("expected en passant notation")
	}

	mustMove(t, client, "exf6")

	if len(events) != 1 {
		t.Fatalf("expected en passant event")
	}
}

func TestGetFEN(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	if got := client.getFEN(); got != "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR" {
		t.Fatalf("unexpected FEN %s", got)
	}
}

func TestFromFENRespectsSideToMove(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1"
	client, err := createAlgebraicGameClientFromFEN(fen, algebraicClientOptions{})
	if err != nil {
		t.Fatalf("fromFEN failed: %v", err)
	}

	expectedFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"
	if got := client.getFEN(); got != expectedFEN {
		t.Fatalf("board mismatch: got %s want %s", got, expectedFEN)
	}

	if _, err := client.move("e4", false); err == nil {
		t.Fatalf("expected white move to fail")
	}

	res := mustMove(t, client, "e5")
	if res.Move.PostSquare.File != 'e' || res.Move.PostSquare.Rank != 5 {
		t.Fatalf("expected move to e5")
	}
	if res.Move.PostSquare.Piece == nil || res.Move.PostSquare.Piece.Side != sideBlack {
		t.Fatalf("expected black piece")
	}
}

func TestIssue71UndoRestoresStatus(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "c5").Undo()

	status := mustStatus(t, client, false)
	if status.Board.LastMovedPiece == nil || status.Board.LastMovedPiece.Side != sideWhite {
		t.Fatalf("expected last moved piece to be white")
	}
	if _, ok := status.NotatedMoves["c5"]; !ok {
		t.Fatalf("expected c5 in notated moves")
	}
}

func TestIssue77UndoFirstMove(t *testing.T) {
	client := createAlgebraicGameClient(algebraicClientOptions{})

	mustMove(t, client, "e4").Undo()

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["e4"]; !ok {
		t.Fatalf("expected e4 available after undo")
	}
}
