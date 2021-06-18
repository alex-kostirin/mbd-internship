package aggregate

import (
	"github.com/pkg/errors"
	"math/big"
)

type Avg struct {
	dividend *big.Float
	divisor  *big.Int
}

func NewAvg() *Avg {
	return &Avg{
		dividend: big.NewFloat(0),
		divisor:  big.NewInt(0),
	}
}

func (a *Avg) Add(data float64) {
	a.dividend.Add(a.dividend, big.NewFloat(data))
	a.divisor.Add(a.divisor, big.NewInt(1))
}

func (a *Avg) Count() float64 {
	res, _ := new(big.Float).Quo(a.dividend, new(big.Float).SetInt(a.divisor)).Float64()
	return res
}

func (a *Avg) Read(p []byte) (int, error) {
	n := 0
	if len(p) == 0 {
		return 0, errors.New("empty bytes")
	}
	dividendLen := int(p[0])
	n++
	if len(p[n:]) < dividendLen {
		return 0, errors.New("insufficient bytes for dividend")
	}
	if err := a.dividend.GobDecode(p[n : n+dividendLen]); err != nil {
		return 0, errors.Wrap(err, "can not decode dividend")
	}
	n += dividendLen
	if len(p[n:]) == 0 {
		return 0, errors.New("divisor len is empty")
	}
	divisorLen := int(p[n])
	n++
	if len(p[n:]) < divisorLen {
		return 0, errors.New("insufficient bytes for divisor")
	}
	if err := a.divisor.GobDecode(p[n : n+divisorLen]); err != nil {
		return 0, errors.Wrap(err, "can not decode dividend")
	}
	n += divisorLen
	return n, nil
}

func (a *Avg) Encode() ([]byte, error) {
	dividendBytes, err := a.dividend.GobEncode()
	if err != nil {
		return nil, errors.Wrap(err, "can not encode dividend")
	}
	divisorBytes, err := a.divisor.GobEncode()
	if err != nil {
		return nil, errors.Wrap(err, "can not encode divisor")
	}

	n := 1 + len(dividendBytes) + 1 + len(divisorBytes)
	res := make([]byte, n)
	cursor := 0
	res[cursor] = byte(len(dividendBytes))
	cursor++
	copy(res[cursor:], dividendBytes)
	cursor += len(dividendBytes)
	res[cursor] = byte(len(divisorBytes))
	cursor++
	copy(res[cursor:], divisorBytes)
	return res, nil
}
