package entity

type Entity interface {
}

type Interactor interface {

	// Interact target is the entity we left-clicked or are interacting with
	Interact(target Entity)
}

type Attacker interface {
	Attack(a Entity)
}
