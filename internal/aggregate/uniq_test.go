package aggregate

import (
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestUniqHyperLogLog256_Add_Count(t *testing.T) {
	realUniq := map[string]struct{}{}
	uniq := NewUniqHyperLogLog256()
	space := uuid.MustParse("69a76476-bbb1-11eb-8529-0242ac130003")
	rand.Seed(42)

	for i := 0; i < 100000; i++ {
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(rand.Intn(300)))
		id := uuid.NewSHA1(space, b)

		realUniq[id.String()] = struct{}{}
		uniq.Add(id[:])
	}

	require.InDelta(t, len(realUniq), uniq.Count(), 0.1*float64(len(realUniq)))
}

func TestUniqHyperLogLog256_Read_Write(t *testing.T) {
	uniq := NewUniqHyperLogLog256()
	space := uuid.MustParse("69a76476-bbb1-11eb-8529-0242ac130003")
	rand.Seed(42)

	for i := 0; i < 100000; i++ {
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(rand.Intn(math.MaxUint16)))
		id := uuid.NewSHA1(space, b)
		uniq.Add(id[:])
	}
	expectedCount := uniq.Count()

	p, err := uniq.Encode()
	require.NoError(t, err)
	uniqFromBytes := NewUniqHyperLogLog256()
	_, err = uniqFromBytes.Read(p)
	require.NoError(t, err)
	require.Equal(t, expectedCount, uniqFromBytes.Count())
}
