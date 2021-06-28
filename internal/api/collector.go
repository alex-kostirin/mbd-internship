package api

import (
	"encoding/json"
	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"math/rand"
	"mbd-internship/internal/geo"
	"net/http"
	"sync"
	"time"
)

type CollectorRequest struct {
	Lat    float64   `json:"lat"`
	Lon    float64   `json:"lon"`
	UserID uuid.UUID `json:"user_id"`
	Signal float64   `json:"signal"`
}

func collector(worker *collectorWorker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		requestID := rand.Int()
		request := CollectorRequest{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Infof("%d: can not decode request: %+v", requestID, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		coords := geo.DecCoords{
			Lat: request.Lat,
			Lon: request.Lon,
		}
		if !coords.IsValid() {
			log.Infof("%d: decemical coordinates are not valid", requestID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !(request.Signal >= 0 && request.Signal <= 100) {
			log.Infof("%d: signal are not valid", requestID)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		worker.process(coords, request.UserID, request.Signal)
		w.WriteHeader(http.StatusOK)
		return
	}
}

type collectorWorker struct {
	db    *leveldb.DB
	queue chan queueItem
	done  sync.WaitGroup

	pc int
	sc int
	ec int
}

type queueItem struct {
	coords geo.DecCoords
	userID uuid.UUID
	signal float64
}

func newCollectorWorker(db *leveldb.DB, queueSize int) *collectorWorker {
	return &collectorWorker{db: db, queue: make(chan queueItem, queueSize)}
}

func (w *collectorWorker) process(coords geo.DecCoords, userID uuid.UUID, signal float64) {
	item := queueItem{coords: coords, userID: userID, signal: signal}
	select {
	case w.queue <- item:
	default:
		w.sc++
		log.Debugf("skipped item %v", item)
	}
}

func (w *collectorWorker) start() {
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		w.done.Add(1)
		lastTickTime := time.Now()
		for item := range w.queue {
			select {
			case tickTime := <-ticker.C:
				pps := float64(w.pc) / (float64(tickTime.Sub(lastTickTime)) / float64(time.Second))
				sps := float64(w.sc) / (float64(tickTime.Sub(lastTickTime)) / float64(time.Second))
				eps := float64(w.ec) / (float64(tickTime.Sub(lastTickTime)) / float64(time.Second))
				log.Infof("worker PPS: %.f SPS: %.f EPS: %.f", pps, sps, eps)
				lastTickTime = tickTime
				w.pc = 0
				w.sc = 0
				w.ec = 0
			default:
			}
			cellID, err := geo.S2CellIDFromDecCoords(item.coords)
			if err != nil {
				log.Debugf("failed get cell ID from coordinates %+v", errors.WithStack(err))
				w.ec++
				continue
			}
			err = addStateDB(w.db, cellID, item.userID, item.signal)
			if err != nil {
				log.Debugf("failed add state to db %+v", errors.WithStack(err))
				w.ec++
				continue
			}
			w.pc++
		}
		w.done.Done()
	}()
}

func (w *collectorWorker) stop() {
	close(w.queue)
	w.done.Wait()
}

func addStateDB(db *leveldb.DB, cellID s2.CellID, userID uuid.UUID, signal float64) error {
	key := s2CellIdToKey(cellID)
	value, err := db.Get(key, nil)
	state := NewCellState()
	switch err {
	case leveldb.ErrNotFound:
	case nil:
		if _, err = state.Read(value); err != nil {
			return errors.Wrap(err, "can not read state")
		}
	default:
		return errors.Wrap(err, "can not read from db")
	}
	state.Add(userID, signal)
	value, err = state.Encode()
	if err != nil {
		return errors.Wrap(err, "can not encode state")
	}
	if err := db.Put(key, value, nil); err != nil {
		return errors.Wrap(err, "can not write state")
	}
	return nil
}
