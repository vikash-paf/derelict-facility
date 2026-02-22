package ecs

type Entity uint32

type Component interface{}

type World struct {
	nextEntityID Entity

	componentStores map[string]map[Entity]Component
}

func NewWorld() *World {
	return &World{
		nextEntityID:    1,
		componentStores: make(map[string]map[Entity]Component),
	}
}

func (w *World) CreateEntity() Entity {
	id := w.nextEntityID
	w.nextEntityID++
	return id
}

func (w *World) DestroyEntity(e Entity) {
	for _, store := range w.componentStores {
		delete(store, e)
	}
}

func (w *World) AddComponent(e Entity, name string, c Component) {
	store, exists := w.componentStores[name]
	if !exists {
		store = make(map[Entity]Component)
		w.componentStores[name] = store
	}
	store[e] = c
}

func (w *World) GetComponent(e Entity, name string) Component {
	store, exists := w.componentStores[name]
	if !exists {
		return nil
	}
	return store[e]
}

func (w *World) RemoveComponent(e Entity, name string) {
	store, exists := w.componentStores[name]
	if exists {
		delete(store, e)
	}
}

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
