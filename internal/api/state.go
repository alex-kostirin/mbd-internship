package api

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"mbd-internship/internal/aggregate"
)

type CellState struct {
	uniq *aggregate.UniqHyperLogLog256
	avg  *aggregate.Avg
}

func NewCellState() *CellState {
	return &CellState{
		uniq: aggregate.NewUniqHyperLogLog256(),
		avg:  aggregate.NewAvg(),
	}
}

func (c *CellState) Add(userID uuid.UUID, avg float64) {
	c.uniq.Add(userID[:])
	c.avg.Add(avg)
}

func (c *CellState) Count() (uniq uint64, avg float64) {
	return c.uniq.Count(), c.avg.Count()
}

func (c *CellState) Read(p []byte) (int, error) {
	n := 0
	rn, err := c.uniq.Read(p)
	if err != nil {
		return n, errors.WithStack(err)
	}
	n += rn
	rn, err = c.avg.Read(p[n:])
	if err != nil {
		return n, errors.WithStack(err)
	}
	n += rn
	return n, nil
}

func (c *CellState) Encode() ([]byte, error) {
	uniqBytes, err := c.uniq.Encode()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	avgBytes, err := c.avg.Encode()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return append(uniqBytes, avgBytes...), nil
}
