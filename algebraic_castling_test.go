package chess

import "testing"

func TestWhiteCastleLeftEvent(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	var events []*MoveEvent

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*MoveEvent); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.GetSquare('b', 1).Piece = nil
	client.game.Board.GetSquare('c', 1).Piece = nil
	client.game.Board.GetSquare('d', 1).Piece = nil

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{PGN: true})
	var events []*MoveEvent

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*MoveEvent); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.GetSquare('b', 1).Piece = nil
	client.game.Board.GetSquare('c', 1).Piece = nil
	client.game.Board.GetSquare('d', 1).Piece = nil

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	var events []*MoveEvent

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*MoveEvent); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.GetSquare('f', 8).Piece = nil
	client.game.Board.GetSquare('g', 8).Piece = nil
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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{PGN: true})
	var events []*MoveEvent

	client.On("castle", func(data interface{}) {
		if mv, ok := data.(*MoveEvent); ok {
			events = append(events, mv)
		}
	})

	client.game.Board.GetSquare('f', 8).Piece = nil
	client.game.Board.GetSquare('g', 8).Piece = nil
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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	client.game.Board.GetSquare('b', 1).Piece = nil
	client.game.Board.GetSquare('c', 1).Piece = nil
	client.game.Board.GetSquare('d', 1).Piece = nil
	mustStatus(t, client, true)

	res := mustMove(t, client, "O-O-O")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}

func TestParseWhiteCastleLeftPGN(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{PGN: true})

	client.game.Board.GetSquare('b', 1).Piece = nil
	client.game.Board.GetSquare('c', 1).Piece = nil
	client.game.Board.GetSquare('d', 1).Piece = nil
	mustStatus(t, client, true)

	res := mustMove(t, client, "0-0-0")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}

func TestParseBlackCastleRight(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	client.game.Board.GetSquare('f', 8).Piece = nil
	client.game.Board.GetSquare('g', 8).Piece = nil
	mustStatus(t, client, true)
	mustMove(t, client, "a4")

	res := mustMove(t, client, "O-O")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}

func TestParseBlackCastleRightPGN(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	client.game.Board.GetSquare('f', 8).Piece = nil
	client.game.Board.GetSquare('g', 8).Piece = nil
	mustStatus(t, client, true)
	mustMove(t, client, "a4")

	res := mustMove(t, client, "0-0")
	if !res.Move.Castle {
		t.Fatalf("expected castle move")
	}
}
