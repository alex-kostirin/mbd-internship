package aggregate

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestAvg_Add_Count(t *testing.T) {
	dividend := 0.0
	divisor := 0
	avg := NewAvg()
	rand.Seed(42)

	for i := 0; i < 100000; i++ {
		add := rand.Float64() * 100
		dividend += add
		divisor++
		avg.Add(add)
	}
	expected := dividend / float64(divisor)
	require.InDelta(t, expected, avg.Count(), expected*0.01)
}

func TestAvg_Read_Write(t *testing.T) {
	avg := NewAvg()
	rand.Seed(42)

	for i := 0; i < 100000; i++ {
		add := rand.Float64() * 100
		avg.Add(add)
	}
	expectedCount := avg.Count()

	p, err := avg.Encode()
	require.NoError(t, err)
	avgFromBytes := NewAvg()
	_, err = avgFromBytes.Read(p)
	require.NoError(t, err)
	require.Equal(t, expectedCount, avgFromBytes.Count())
}
