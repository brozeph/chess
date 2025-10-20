package chess

type Side int

const (
	sideWhite Side = iota
	sideBlack
)

/*
func (s Side) name() string {
	if s == sideWhite {
		return "white"
	}
	return "black"
}
*/

func (s Side) opponent() Side {
	if s == sideWhite {
		return sideBlack
	}
	return sideWhite
}
