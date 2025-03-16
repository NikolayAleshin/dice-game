package service

import "dice-game/pkg/infrastructure/random"

type RandomServiceInterface interface {
	GetRandomGenerator() (random.Generator, error)
	GetGeneratorByName(name string) (random.Generator, error)
	AddGenerator(generator random.Generator)
}

type VerifiableGenerator interface {
	random.Generator
	GetVerificationData() string
}
