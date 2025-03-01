package usecase

type repository interface {
}

type Processor struct {
	repo repository
}

func NewProcessor(repo repository) *Processor {
	return &Processor{repo: repo}
}
