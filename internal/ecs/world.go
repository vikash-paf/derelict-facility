package ecs

// Entity is conceptually just an ID.
// It represents a single "thing" in the game world, but contains no data itself.
type Entity uint32

// Component is a marker interface.
// Any struct can be a Component as long as it contains pure data and no logic.
type Component interface{}

// World is the central registry for the Entity-Component System.
// It manages the creation of Entities and the storage of Components.
type World struct {
	nextEntityID Entity

	// componentStores maps a component's name (string) to a map of Entity -> Component
	// E.g., "Position" -> { 1: Position{X:5, Y:5}, 2: Position{X:10, Y:10} }
	componentStores map[string]map[Entity]Component
}

// NewWorld creates and initializes a new ECS World.
func NewWorld() *World {
	return &World{
		nextEntityID:    1, // 0 can be reserved for "null" entity
		componentStores: make(map[string]map[Entity]Component),
	}
}

// CreateEntity allocates and returns a new unique Entity ID.
func (w *World) CreateEntity() Entity {
	id := w.nextEntityID
	w.nextEntityID++
	return id
}

// DestroyEntity removes an entity and all of its associated components from the world.
func (w *World) DestroyEntity(e Entity) {
	for _, store := range w.componentStores {
		delete(store, e)
	}
}

// AddComponent attaches a component struct to an entity.
// We use the string name of the component (e.g., "Position") to store it in the right bucket.
func (w *World) AddComponent(e Entity, name string, c Component) {
	store, exists := w.componentStores[name]
	if !exists {
		store = make(map[Entity]Component)
		w.componentStores[name] = store
	}
	store[e] = c
}

// GetComponent retrieves a component interface for a given entity and component name.
// Returns nil if the entity does not have that component.
func (w *World) GetComponent(e Entity, name string) Component {
	store, exists := w.componentStores[name]
	if !exists {
		return nil
	}
	return store[e]
}

// RemoveComponent removes a specific component from an entity.
func (w *World) RemoveComponent(e Entity, name string) {
	store, exists := w.componentStores[name]
	if exists {
		delete(store, e)
	}
}

// GetEntitiesWith returns a list of all Entity IDs that possess a specific component.
// This is the backbone of how Systems query the World.
func (w *World) GetEntitiesWith(componentName string) []Entity {
	store, exists := w.componentStores[componentName]
	if !exists {
		return nil
	}

	entities := make([]Entity, 0, len(store))
	for e := range store {
		entities = append(entities, e)
	}
	return entities
}
