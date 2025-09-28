package random

import (
	"github.com/stretchr/testify/mock"
)

type MockRandom struct { 
	mock.Mock 
}

func (m *MockRandom) Float64() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}
func (m *MockRandom) Intn(n int) int {
	args := m.Called(n)
	return args.Int(0)
}
