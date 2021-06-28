package main

import (
	"encoding/binary"
	"github.com/google/uuid"
	"math"
	"math/rand"
)

type Config struct {
	LatRange      [2]float64
	LonRange      [2]float64
	AnomalyPoints [][2]float64

	generateAnomaly       bool
	lastAnomalyPointIndex int
}

func randFromRange(v1, v2 float64) float64 {
	return v1 + rand.Float64()*(v2-v1)
}

func (c *Config) LatLon() (lat float64, lon float64) {
	if len(c.AnomalyPoints) > 0 && c.generateAnomaly {
		anomalyPoint := c.AnomalyPoints[c.lastAnomalyPointIndex]
		lat = randFromRange(anomalyPoint[0], anomalyPoint[0]+0.005)
		lon = randFromRange(anomalyPoint[1], anomalyPoint[1]+0.005)
		c.lastAnomalyPointIndex = (c.lastAnomalyPointIndex + 1) % len(c.AnomalyPoints)
	} else {
		lat = randFromRange(c.LatRange[0], c.LatRange[1])
		lon = randFromRange(c.LonRange[0], c.LonRange[1])
	}
	c.generateAnomaly = !c.generateAnomaly
	return lat, lon
}

var space = uuid.MustParse("69a76476-bbb1-11eb-8529-0242ac130003")

func (c *Config) UserID() uuid.UUID {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(rand.Intn(math.MaxUint16)))
	return uuid.NewSHA1(space, b)
}

func (c *Config) Signal() float64 {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(rand.Intn(math.MaxUint16)))
	return rand.Float64() * 100.0
}
