package random

type RandomProvider interface {
	Float64() float64
	Intn(n int) int
}