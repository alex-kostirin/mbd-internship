package api

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
)

type Handler struct {
	*http.ServeMux
	db     *leveldb.DB
	worker *collectorWorker
}

func NewHandler() (*Handler, error) {
	db, err := leveldb.OpenFile("./db.leveldb", nil)
	if err != nil {
		return nil, errors.Wrap(err, "can not create db")
	}
	worker := newCollectorWorker(db, 10000)
	mux := http.NewServeMux()
	mux.HandleFunc("/collector", collector(worker))
	mux.HandleFunc("/data", data(db))

	h := &Handler{
		ServeMux: mux,
		db:       db,
		worker:   worker,
	}
	worker.start()
	return h, nil
}

func (h *Handler) Stop() {
	h.worker.stop()
	h.db.Close()
}
