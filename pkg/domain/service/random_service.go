package service

import (
	"dice-game/pkg/infrastructure/random"
	"errors"
	"math/rand"
	"time"
)

type RandomService struct {
	generators []random.Generator
	rnd        *rand.Rand
}

func NewRandomService(generators []random.Generator) *RandomService {
	return &RandomService{
		generators: generators,
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *RandomService) GetRandomGenerator() (random.Generator, error) {
	if len(s.generators) == 0 {
		return nil, errors.New("no random generators available")
	}

	idx := s.rnd.Intn(len(s.generators))

	return s.generators[idx], nil
}

func (s *RandomService) GetGeneratorByName(name string) (random.Generator, error) {
	for _, gen := range s.generators {
		if gen.Name() == name {
			return gen, nil
		}
	}

	return nil, errors.New("generator not found")
}

func (s *RandomService) AddGenerator(generator random.Generator) {
	s.generators = append(s.generators, generator)
}
