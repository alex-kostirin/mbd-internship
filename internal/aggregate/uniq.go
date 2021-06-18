package aggregate

import (
	"github.com/pkg/errors"
	"hash/fnv"
	"math"
	"math/bits"
)

const b = 8
const m = 256
const alpha = 0.7213 / (1 + (1.079 / m))

type UniqHyperLogLog256 struct {
	state [m]uint8
}

func NewUniqHyperLogLog256() *UniqHyperLogLog256 {
	return new(UniqHyperLogLog256)
}

func (u *UniqHyperLogLog256) Add(data []byte) {
	x := h(data)
	j := x >> (32 - b)
	rank := p(x)

	if rank > u.state[j] {
		u.state[j] = rank
	}
}

func (u *UniqHyperLogLog256) Count() uint64 {
	sum := 0.0
	for _, rank := range u.state {
		sum += math.Pow(2, -float64(rank))
	}
	e := alpha * m * m / sum
	return uint64(e)
}

func (u *UniqHyperLogLog256) Read(p []byte) (int, error) {
	if len(p) < m {
		return 0, errors.Errorf("byte slice must contain at least %d bytes", m)
	}
	copy(u.state[:], p[:m])
	return m, nil
}

func (u *UniqHyperLogLog256) Encode() ([]byte, error) {
	res := make([]byte, m)
	copy(res, u.state[:])
	return res, nil
}

func h(v []byte) uint32 {
	h := fnv.New32()
	h.Write(v)
	return h.Sum32()
}

func p(w uint32) uint8 {
	return uint8(1 + bits.TrailingZeros32(w))
}
