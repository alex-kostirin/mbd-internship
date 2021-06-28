package main

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"sync/atomic"
)

type HTTPWorker struct {
	client   *RLHTTPClient
	jobs     chan *http.Request
	counters *Counters
	wg       *sync.WaitGroup
}

func NewHTTPWorker(rateLimiter *rate.Limiter, disableKeepAlives bool, jobs chan *http.Request, counters *Counters, wg *sync.WaitGroup) *HTTPWorker {
	return &HTTPWorker{
		client:   NewRLHTTPClient(rateLimiter, disableKeepAlives),
		jobs:     jobs,
		counters: counters,
		wg:       wg,
	}
}

func (w *HTTPWorker) Run() {
	w.wg.Add(1)
	for job := range w.jobs {
		atomic.AddInt64(&w.counters.Total, 1)
		if _, err := w.client.Do(job); err != nil {
			atomic.AddInt64(&w.counters.Errors, 1)
			log.Debugf("request error: %+v", err)
			continue
		}
		atomic.AddInt64(&w.counters.Completed, 1)
	}
	w.wg.Done()
}

type HTTPWorkerPool struct {
	jobs    chan *http.Request
	workers []*HTTPWorker
	wg      *sync.WaitGroup
}

func NewHTTPWorkerPool(size int, rateLimiter *rate.Limiter, disableKeepAlives bool, counters *Counters) *HTTPWorkerPool {
	jobs := make(chan *http.Request, size)
	wg := sync.WaitGroup{}
	pool := &HTTPWorkerPool{jobs: jobs, wg: &wg}
	var workers []*HTTPWorker
	for i := 0; i < size; i++ {
		worker := NewHTTPWorker(rateLimiter, disableKeepAlives, pool.jobs, counters, pool.wg)
		workers = append(workers, worker)
	}
	pool.workers = workers
	return pool
}

func (p *HTTPWorkerPool) Start() {
	for _, worker := range p.workers {
		go worker.Run()
	}
}

func (p *HTTPWorkerPool) ProcessRequest(request *http.Request) {
	p.jobs <- request
}

func (p *HTTPWorkerPool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}
