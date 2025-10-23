package chess

import "testing"

func TestMoveEventTriggered(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	var events []*moveEvent

	client.On("move", func(data interface{}) {
		if mv, ok := data.(*moveEvent); ok {
			events = append(events, mv)
		}
	})

	mustMove(t, client, "b4")
	mustMove(t, client, "e6")

	if len(events) != 2 {
		t.Fatalf("expected 2 move events, got %d", len(events))
	}
}

func TestCaptureEvent(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	var events []*moveEvent

	client.On("capture", func(data interface{}) {
		if mv, ok := data.(*moveEvent); ok {
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

func TestPromotionEvent(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	var events []*Square

	client.On("promote", func(data interface{}) {
		if sq, ok := data.(*Square); ok {
			events = append(events, sq)
		}
	})

	for _, sq := range []string{"a7", "a8", "b8", "c8", "d8", "a2"} {
		client.game.Board.getSquareByName(sq).Piece = nil
	}
	client.game.Board.GetSquare('a', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.GetSquare('a', 7).Piece.MoveCount = 1

	mustStatus(t, client, true)
	mustMove(t, client, "a8R")

	if len(client.game.MoveHistory) == 0 || !client.game.MoveHistory[0].Promotion {
		t.Fatalf("expected promotion flag")
	}
	if len(events) != 1 {
		t.Fatalf("expected promotion event")
	}
}

func TestIssue23CheckEvent(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	var event kingThreatEvent
	triggered := false

	client.On("check", func(data interface{}) {
		if ev, ok := data.(kingThreatEvent); ok {
			event = ev
			triggered = true
		}
	})

	client.game.Board.GetSquare('b', 1).Piece = nil
	client.game.Board.GetSquare('f', 6).Piece = newPiece(pieceKnight, sideWhite)
	client.game.Board.GetSquare('f', 6).Piece.MoveCount = 1

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

func TestIssue53EnPassantEvent(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	var events []*moveEvent

	client.On("enPassant", func(data interface{}) {
		if mv, ok := data.(*moveEvent); ok {
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
