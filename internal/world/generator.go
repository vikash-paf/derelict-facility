package world

type MapGenerator interface {
	Generate(width, height int) (*Map, int, int) // Returns the generated map and player position (x,y)
}
