package api

import (
	"encoding/json"
	"github.com/golang/geo/s2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"math/rand"
	"mbd-internship/internal/geo"
	"net/http"
)

type dataRequest struct {
	Area []geo.DecCoords `json:"area"`
}

type dataResponse []dataResponseItem

type dataResponseItem struct {
	S2ID          uint64          `json:"s2_id"`
	S2Coordinates []geo.DecCoords `json:"s2_coordinates"`
	UniqUsers     uint64          `json:"uniq_users"`
	SignalAvg     float64         `json:"signal_avg"`
}

func data(db *leveldb.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		requestID := rand.Int()
		request := dataRequest{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Infof("%d: can not decode request: %s", requestID, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		requestBody, _ := json.Marshal(request)
		log.Debugf("%d: data request, body: %s", requestID, string(requestBody))
		for _, dc := range request.Area {
			if !dc.IsValid() {
				log.Infof("%d: decemical coordinates are not valid", requestID)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		cellUnion, err := geo.CoverRegionS2CellUnion(request.Area[0], request.Area[1])
		if err != nil {
			log.Infof("%d: can not convert region to cell union: %+v", requestID, errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debugf("%d: got %d cells", requestID, len(cellUnion))
		states, err := getStatesDB(db, cellUnion)
		if err != nil {
			log.Infof("%d: can not get state from DB: %+v", requestID, errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		response := dataResponse{}
		for cellID, state := range states {
			coords, err := geo.DecCoordsFromS2CellID(cellID)
			if err != nil {
				log.Infof("%d: can not convert dec coords to cell ID: %+v", requestID, errors.WithStack(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			uniqUsers, signalAvg := state.Count()
			response = append(response, dataResponseItem{
				S2ID:          uint64(cellID),
				S2Coordinates: coords,
				UniqUsers:     uniqUsers,
				SignalAvg:     signalAvg,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Infof("%d: can not encode response: %+v", requestID, errors.WithStack(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
}

func getStatesDB(db *leveldb.DB, cellUnion s2.CellUnion) (map[s2.CellID]*CellState, error) {
	snapshot, err := db.GetSnapshot()
	if err != nil {
		return nil, errors.Wrap(err, "can not get snapshot")
	}
	defer snapshot.Release()

	res := map[s2.CellID]*CellState{}
	for _, cellID := range cellUnion {
		key := s2CellIdToKey(cellID)
		value, err := snapshot.Get(key, nil)
		switch err {
		case leveldb.ErrNotFound:
		case nil:
			state := NewCellState()
			if _, err := state.Read(value); err != nil {
				return nil, errors.Wrap(err, "can not read state")
			}
			res[cellID] = state
		default:
			return nil, errors.Wrap(err, "can not read from db")
		}
	}
	return res, nil
}
