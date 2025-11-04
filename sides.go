package chess

// Side represents a player's color (White or Black).
type Side int

const (
	// sideWhite represents the white side.
	sideWhite Side = iota
	// sideBlack represents the black side.
	sideBlack
)

// Name returns the string representation of the side ("white" or "black").
func (s Side) Name() string {
	if s == sideWhite {
		return "white"
	}
	return "black"
}

// Opponent returns the opposing side.
func (s Side) Opponent() Side {
	if s == sideWhite {
		return sideBlack
	}
	return sideWhite
}
