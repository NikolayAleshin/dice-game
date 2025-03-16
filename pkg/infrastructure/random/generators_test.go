package random

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestStandardGenerator_Generate(t *testing.T) {
	g := NewStandardGenerator()

	for i := 0; i < 1000; i++ {
		val, err := g.Generate(1, 6)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, val, 1)
		assert.LessOrEqual(t, val, 6)
	}

	val, err := g.Generate(6, 1)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, val, 1)
	assert.LessOrEqual(t, val, 6)
}

func TestCryptoGenerator_Generate(t *testing.T) {
	g := NewCryptoGenerator()

	for i := 0; i < 1000; i++ {
		val, err := g.Generate(1, 6)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, val, 1)
		assert.LessOrEqual(t, val, 6)
	}

	val, err := g.Generate(6, 1)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, val, 1)
	assert.LessOrEqual(t, val, 6)
}

func TestProovablyFairGenerator_Generate(t *testing.T) {
	g := NewProovablyFairGenerator("serverSeed", func() string {
		return "clientSeed"
	})

	for i := 0; i < 1000; i++ {
		val, err := g.Generate(1, 6)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, val, 1)
		assert.LessOrEqual(t, val, 6)
	}

	_, err := g.Generate(6, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min cannot be greater than max")
}

func TestProovablyFairGenerator_GetVerificationData(t *testing.T) {
	serverSeed := "test-server-seed"
	clientSeed := "test-client-seed"

	g := NewProovablyFairGenerator(serverSeed, func() string {
		return clientSeed
	})

	_, err := g.Generate(1, 6)
	assert.NoError(t, err)

	verificationData := g.GetVerificationData()

	assert.Contains(t, verificationData, serverSeed)
	assert.Contains(t, verificationData, ":1:")

	_, err = g.Generate(1, 6)
	assert.NoError(t, err)

	newVerificationData := g.GetVerificationData()
	assert.Contains(t, newVerificationData, ":2:")
	assert.NotEqual(t, verificationData, newVerificationData)
}

func TestGenerators_Concurrency(t *testing.T) {
	g := NewStandardGenerator()

	const goroutines = 10
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				val, err := g.Generate(1, 6)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, val, 1)
				assert.LessOrEqual(t, val, 6)
			}
		}()
	}

	wg.Wait()
}

func TestGenerators_Name(t *testing.T) {
	standardGen := NewStandardGenerator()
	assert.Equal(t, "standard", standardGen.Name())

	cryptoGen := NewCryptoGenerator()
	assert.Equal(t, "crypto", cryptoGen.Name())

	provablyFairGen := NewProovablyFairGenerator("seed", func() string { return "client" })
	assert.Equal(t, "provably_fair", provablyFairGen.Name())
}
