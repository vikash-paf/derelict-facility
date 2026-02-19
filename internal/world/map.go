package world

type Map struct {
	Width  int
	Height int

	// keeping this as a single slice for now,
	// this will store the location, it's continuous "tape", better performance and low memory usage
	Tiles []Tile
}

func NewMap(width, height int) *Map {
	return &Map{
		Width:  width,
		Height: height,
		Tiles:  make([]Tile, width*height),
	}
}
