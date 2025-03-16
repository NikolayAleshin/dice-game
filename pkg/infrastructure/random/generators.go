package random

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"sync"
	"time"
)

type Generator interface {
	Generate(min, max int) (int, error)
	Name() string
}

type StandardGenerator struct {
	source *mathrand.Rand
	mu     sync.Mutex
}

func NewStandardGenerator() *StandardGenerator {
	return &StandardGenerator{
		source: mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
	}
}

func (g *StandardGenerator) Generate(min, max int) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if min > max {
		min, max = max, min
	}

	return g.source.Intn(max-min+1) + min, nil
}

func (g *StandardGenerator) Name() string {
	return "standard"
}

type CryptoGenerator struct{}

func NewCryptoGenerator() *CryptoGenerator {
	return &CryptoGenerator{}
}

func (g *CryptoGenerator) Generate(min, max int) (int, error) {
	if min > max {
		min, max = max, min
	}

	delta := max - min + 1

	n, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return 0, err
	}

	return int(n.Int64()) + min, nil
}

func (g *CryptoGenerator) Name() string {
	return "crypto"
}

type ProovablyFairGenerator struct {
	serverSeed     string
	clientSeedFunc func() string
	nonce          int
}

func NewProovablyFairGenerator(serverSeed string, clientSeedFunc func() string) *ProovablyFairGenerator {
	return &ProovablyFairGenerator{
		serverSeed:     serverSeed,
		clientSeedFunc: clientSeedFunc,
		nonce:          0,
	}
}

func (g *ProovablyFairGenerator) Generate(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}

	clientSeed := g.clientSeedFunc()
	g.nonce++

	combined := fmt.Sprintf("%s:%s:%d", g.serverSeed, clientSeed, g.nonce)

	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hexPart := hash[:8]
	num, err := hex.DecodeString(hexPart)
	if err != nil {
		return 0, fmt.Errorf("failed to decode hex: %w", err)
	}

	var value int = 0
	for _, b := range num {
		value = (value << 8) | int(b)
	}

	rangeSize := max - min + 1
	result := (value % rangeSize) + min

	return result, nil
}

func (g *ProovablyFairGenerator) Name() string {
	return "provably_fair"
}

func (g *ProovablyFairGenerator) GetVerificationData() string {
	clientSeed := g.clientSeedFunc()

	combined := fmt.Sprintf("%s:%s:%d", g.serverSeed, clientSeed, g.nonce)

	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hash := hex.EncodeToString(hasher.Sum(nil))

	return fmt.Sprintf("%s:%d:%s", g.serverSeed, g.nonce, hash)
}
