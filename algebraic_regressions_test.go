package chess

import "testing"

func TestIssue1NoPhantomPawn(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	target := client.game.Board.GetSquare('c', 5)

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	target := client.game.Board.GetSquare('a', 6)

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})
	target := client.game.Board.GetSquare('e', 6)

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
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

	mustMove(t, client, "e4")
	mustMove(t, client, "a5")
	mustMove(t, client, "Ba6")

	status := mustStatus(t, client, false)
	if _, ok := status.NotatedMoves["b5"]; !ok {
		t.Fatalf("expected pawn to advance two squares")
	}
}

func TestIssue43ParseGxf3Check(t *testing.T) {
	client := CreateAlgebraicGameClient(AlgebraicClientOptions{})

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
