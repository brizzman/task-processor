package random

import (
	"crypto/rand"
	"math/big"

	"task-processor/internal/application/ports/inbound/random"
)

type CryptoRandomProvider struct{}

func NewCryptoRandomProvider() random.RandomProvider {
	return &CryptoRandomProvider{}
}

func (rp *CryptoRandomProvider) Float64() float64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	return float64(n.Int64()) / 1_000_000_000
}

func (rp *CryptoRandomProvider) Intn(n int) int {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(num.Int64())
}
