package entity

// Entity represents an actor in the game, such as the player or drones
type Entity struct {
	Name string
	ID   int
	X    int
	Y    int
}

// New creates a new entity
func New(id int, name string, x, y int) *Entity {
	return &Entity{
		ID:   id,
		Name: name,
		X:    x,
		Y:    y,
	}
}
