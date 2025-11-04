package chess

import "testing"

func TestWhitePawnPromotionMoves(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	client.game.Board.GetSquare('a', 7).Piece = nil
	client.game.Board.GetSquare('a', 8).Piece = nil
	client.game.Board.GetSquare('a', 2).Piece = nil
	client.game.Board.GetSquare('a', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.GetSquare('a', 7).Piece.MoveCount = 1

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	client.game.Board.GetSquare('a', 2).Piece = nil
	client.game.Board.GetSquare('a', 1).Piece = nil
	client.game.Board.GetSquare('a', 7).Piece = nil
	client.game.Board.GetSquare('a', 2).Piece = newPiece(piecePawn, sideBlack)
	client.game.Board.GetSquare('a', 2).Piece.MoveCount = 1

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	for _, sq := range []string{"a7", "a8", "b8", "c8", "d8", "a2"} {
		client.game.Board.getSquareByName(sq).Piece = nil
	}
	client.game.Board.GetSquare('a', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.GetSquare('a', 7).Piece.MoveCount = 1

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	for _, sq := range []string{"a2", "a1", "b1", "c1", "d1", "a7"} {
		client.game.Board.getSquareByName(sq).Piece = nil
	}
	client.game.Board.GetSquare('a', 2).Piece = newPiece(piecePawn, sideBlack)
	client.game.Board.GetSquare('a', 2).Piece.MoveCount = 1

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

func TestIssue17PromotionAvailability(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	client.game.Board.GetSquare('c', 7).Piece = nil
	client.game.Board.GetSquare('c', 8).Piece = nil
	client.game.Board.GetSquare('c', 2).Piece = nil
	client.game.Board.GetSquare('c', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.GetSquare('c', 7).Piece.MoveCount = 1
	client.game.Board.GetSquare('h', 7).Piece = nil
	client.game.Board.GetSquare('h', 7).Piece = newPiece(pieceBishop, sideWhite)
	client.game.Board.GetSquare('h', 7).Piece.MoveCount = 1

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	client.game.Board.GetSquare('c', 7).Piece = nil
	client.game.Board.GetSquare('c', 2).Piece = nil
	client.game.Board.GetSquare('c', 7).Piece = newPiece(piecePawn, sideWhite)
	client.game.Board.GetSquare('c', 7).Piece.MoveCount = 1

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
