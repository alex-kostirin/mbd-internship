package geo

import (
	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
)

const defaultLevel = 15

type DecCoords struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (dc DecCoords) IsValid() bool {
	latIsValid := dc.Lat >= -90 && dc.Lat <= 90
	lonIsValid := dc.Lat >= -180 && dc.Lat <= 180
	return latIsValid && lonIsValid
}

func S2CellIDFromDecCoords(dc DecCoords) (s2.CellID, error) {
	if !dc.IsValid() {
		return 0, errors.New("lat lon is not valid")
	}
	ll := s2.LatLngFromDegrees(dc.Lat, dc.Lon)
	cellID := s2.CellIDFromLatLng(ll)
	cellID = cellID.Parent(defaultLevel)
	return cellID, nil
}

func CoverRegionS2CellUnion(sw, ne DecCoords) (s2.CellUnion, error) {
	degrees := [][2]float64{{sw.Lat, sw.Lon}, {ne.Lat, sw.Lon}, {ne.Lat, ne.Lon}, {sw.Lat, ne.Lon}}
	var points []s2.Point
	for _, dg := range degrees {
		ll := s2.LatLngFromDegrees(dg[0], dg[1])
		if !ll.IsValid() {
			return nil, errors.New("lat lon is not valid")
		}
		points = append(points, s2.PointFromLatLng(ll))
	}
	loop := s2.LoopFromPoints(points)
	loop.Invert()
	rc := s2.RegionCoverer{
		MinLevel: defaultLevel,
		MaxLevel: defaultLevel,
		MaxCells: 10000,
	}
	cu := rc.Covering(loop)
	return cu, nil
}

func DecCoordsFromS2CellID(cellID s2.CellID) ([]DecCoords, error) {
	if !cellID.IsValid() {
		return nil, errors.New("cell is not valid")
	}
	cell := s2.CellFromCellID(cellID)
	loop := s2.PolygonFromCell(cell).Loop(0)
	loop.Invert()
	points := loop.Vertices()
	var res []DecCoords
	for _, point := range points {
		ll := s2.LatLngFromPoint(point)
		res = append(res, DecCoords{
			Lat: ll.Lat.Degrees(),
			Lon: ll.Lng.Degrees(),
		})
	}
	return res, nil
}
