package service

import (
	"dice-game/pkg/infrastructure/random"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRandomService(t *testing.T) {
	// Arrange
	gen1 := new(MockGenerator)
	gen2 := new(MockGenerator)
	generators := []random.Generator{gen1, gen2}

	// Act
	service := NewRandomService(generators)

	// Assert
	assert.NotNil(t, service)
	assert.Equal(t, 2, len(service.generators))
	assert.NotNil(t, service.rnd)
}

func TestRandomService_GetRandomGenerator(t *testing.T) {
	t.Run("Success with multiple generators", func(t *testing.T) {
		// Arrange
		gen1 := new(MockGenerator)
		gen2 := new(MockGenerator)
		gen1.On("Name").Return("gen1")
		gen2.On("Name").Return("gen2")
		generators := []random.Generator{gen1, gen2}

		service := NewRandomService(generators)

		// Act
		generator, err := service.GetRandomGenerator()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, generator)
		assert.Contains(t, []random.Generator{gen1, gen2}, generator)
	})

	t.Run("Success with single generator", func(t *testing.T) {
		// Arrange
		gen := new(MockGenerator)
		gen.On("Name").Return("gen")
		generators := []random.Generator{gen}

		service := NewRandomService(generators)

		// Act
		generator, err := service.GetRandomGenerator()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, gen, generator)
	})

	t.Run("Error when no generators", func(t *testing.T) {
		// Arrange
		service := NewRandomService([]random.Generator{})

		// Act
		generator, err := service.GetRandomGenerator()

		// Assert
		assert.Error(t, err)
		assert.Nil(t, generator)
		assert.Equal(t, "no random generators available", err.Error())
	})
}

func TestRandomService_GetGeneratorByName(t *testing.T) {
	t.Run("Success when generator exists", func(t *testing.T) {
		// Arrange
		gen1 := new(MockGenerator)
		gen2 := new(MockGenerator)
		gen1.On("Name").Return("gen1").Maybe()
		gen2.On("Name").Return("gen2").Maybe()
		generators := []random.Generator{gen1, gen2}

		service := NewRandomService(generators)

		// Act
		generator, err := service.GetGeneratorByName("gen2")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, gen2, generator)
	})

	t.Run("Error when generator not found", func(t *testing.T) {
		// Arrange
		gen1 := new(MockGenerator)
		gen1.On("Name").Return("gen1").Maybe()
		generators := []random.Generator{gen1}

		service := NewRandomService(generators)

		// Act
		generator, err := service.GetGeneratorByName("nonexistent")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, generator)
		assert.Equal(t, "generator not found", err.Error())
	})

	t.Run("Error when no generators", func(t *testing.T) {
		// Arrange
		service := NewRandomService([]random.Generator{})

		// Act
		generator, err := service.GetGeneratorByName("any")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, generator)
		assert.Equal(t, "generator not found", err.Error())
	})
}

func TestRandomService_AddGenerator(t *testing.T) {
	// Arrange
	gen1 := new(MockGenerator)
	gen1.On("Name").Return("gen1").Maybe()
	service := NewRandomService([]random.Generator{gen1})
	assert.Equal(t, 1, len(service.generators))

	gen2 := new(MockGenerator)
	gen2.On("Name").Return("gen2").Maybe()

	// Act
	service.AddGenerator(gen2)

	// Assert
	assert.Equal(t, 2, len(service.generators))
	assert.Equal(t, gen2, service.generators[1])

	generator, err := service.GetGeneratorByName("gen2")
	assert.NoError(t, err)
	assert.Equal(t, gen2, generator)
}

func TestRandomService_Deterministic(t *testing.T) {
	t.Run("Deterministic with fixed seed", func(t *testing.T) {
		t.Skip("Skipping deterministic test as current implementation uses time-based seed")

	})
}

func TestRandomService_Distribution(t *testing.T) {
	t.Run("Distribution check", func(t *testing.T) {
		// Arrange
		gen1 := new(MockGenerator)
		gen2 := new(MockGenerator)
		gen3 := new(MockGenerator)
		gen1.On("Name").Return("gen1").Maybe()
		gen2.On("Name").Return("gen2").Maybe()
		gen3.On("Name").Return("gen3").Maybe()
		generators := []random.Generator{gen1, gen2, gen3}

		service := NewRandomService(generators)

		// Act
		const iterations = 1000
		counts := make(map[random.Generator]int)

		for i := 0; i < iterations; i++ {
			gen, err := service.GetRandomGenerator()
			assert.NoError(t, err)
			counts[gen]++
		}

		// Assert
		assert.Equal(t, 3, len(counts))

		expectedCount := iterations / 3
		tolerance := float64(expectedCount) * 0.2

		for _, count := range counts {
			assert.InDelta(t, expectedCount, count, tolerance)
		}
	})
}
