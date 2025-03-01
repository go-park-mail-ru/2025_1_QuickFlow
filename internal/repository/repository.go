package repository

type InMemory struct {
}

func NewInMemory() *InMemory {
    return &InMemory{}
}
