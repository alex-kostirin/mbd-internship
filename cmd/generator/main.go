package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"mbd-internship/internal/api"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var flags = struct {
	debug            bool
	workers          int
	rate             int
	preset           string
	disableKeepAlive bool
}{}

var presets = map[string]*Config{
	"spb": {
		LatRange: [2]float64{59.87, 60.00},
		LonRange: [2]float64{30.19, 30.50},
	},

	"spb-anomaly": {
		LatRange:      [2]float64{59.87, 60.00},
		LonRange:      [2]float64{30.19, 30.50},
		AnomalyPoints: [][2]float64{{60.009, 30.294}, {59.933, 30.314}},
	},

	"world": {
		LatRange: [2]float64{-90, 90},
		LonRange: [2]float64{-180, 180},
	},

	"world-anomaly": {
		LatRange:      [2]float64{-90, 90},
		LonRange:      [2]float64{-180, 180},
		AnomalyPoints: [][2]float64{{60.009, 30.294}, {59.933, 30.314}},
	},

	"errors": {
		LatRange: [2]float64{100, 200},
		LonRange: [2]float64{-300, -200},
	},
}

func init() {
	parseFlags()
}

func parseFlags() {
	flag.BoolVar(&flags.debug, "debug", false, "debug mode")
	flag.IntVar(&flags.workers, "workers", 10, "n workers")
	flag.IntVar(&flags.rate, "rate", 300, "max request rate")
	flag.StringVar(&flags.preset, "preset", "spb-anomaly", "max request rate")
	flag.BoolVar(&flags.disableKeepAlive, "disable-keep-alive", false, "disable keep alive")
	flag.Parse()
	if _, hasPreset := presets[flags.preset]; !hasPreset {
		fmt.Printf(`unknown preset "%s"`, flags.preset)
		flag.Usage()
		os.Exit(2)
	}
}

func main() {
	level := log.InfoLevel
	if flags.debug {
		level = log.DebugLevel
	}
	log.SetLevel(level)

	preset := presets[flags.preset]
	c := &Counters{}
	rl := rate.NewLimiter(rate.Limit(flags.rate), 1)
	wp := NewHTTPWorkerPool(flags.workers, rl, flags.disableKeepAlive, c)
	sigs := make(chan os.Signal, 1)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	lastTickTime := time.Now()

	wp.Start()
	defer wp.Stop()
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case tickTime := <-ticker.C:
			requestsCount := c.Total
			completedCount := c.Completed
			errorsCount := c.Errors
			rps := float64(requestsCount) / (float64(tickTime.Sub(lastTickTime)) / float64(time.Second))
			cps := float64(completedCount) / (float64(tickTime.Sub(lastTickTime)) / float64(time.Second))
			eps := float64(errorsCount) / (float64(tickTime.Sub(lastTickTime)) / float64(time.Second))
			log.Infof("RPS: %.f CPS: %.f EPS: %.f", rps, cps, eps)
			lastTickTime = tickTime
			atomic.AddInt64(&c.Total, -requestsCount)
			atomic.AddInt64(&c.Completed, -completedCount)
			atomic.AddInt64(&c.Errors, -errorsCount)
		case <-sigs:
			return
		default:
		}
		lat, lon := preset.LatLon()
		r := api.CollectorRequest{
			Lat:    lat,
			Lon:    lon,
			UserID: preset.UserID(),
			Signal: preset.Signal(),
		}
		buf, err := json.Marshal(r)
		if err != nil {
			log.Fatalf("%+v", errors.WithStack(err))
		}
		request, err := http.NewRequest("POST", "http://127.0.0.1:8080/collector", bytes.NewBuffer(buf))
		if err != nil {
			log.Fatalf("can not make request %+v", err)
		}
		request.Header.Set("Content-Type", "application/json")
		wp.ProcessRequest(request)
	}
}
