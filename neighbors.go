package chess

type neighbor int

const (
	NeighborAbove            neighbor = 8
	NeighborAboveLeft        neighbor = 7
	NeighborAboveRight       neighbor = 9
	NeighborBelow            neighbor = -8
	NeighborBelowLeft        neighbor = -9
	NeighborBelowRight       neighbor = -7
	NeighborLeft             neighbor = -1
	NeighborRight            neighbor = 1
	NeighborKnightAboveLeft  neighbor = 15
	NeighborKnightAboveRight neighbor = 17
	NeighborKnightBelowLeft  neighbor = -17
	NeighborKnightBelowRight neighbor = -15
	NeighborKnightLeftAbove  neighbor = 6
	NeighborKnightLeftBelow  neighbor = -10
	NeighborKnightRightAbove neighbor = 10
	NeighborKnightRightBelow neighbor = -6
)
