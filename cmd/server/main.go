package main

import (
	"flag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mbd-internship/internal/api"
	"net/http"
	"syscall"
)

var debug bool
var addr string

func init() {
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.StringVar(&addr, "addr", ":8080", "address")
	flag.Parse()
}

func main() {
	level := log.InfoLevel
	if debug {
		level = log.DebugLevel
	}
	log.SetLevel(level)

	apiHandler, err := api.NewHandler()
	if err != nil {
		log.Fatalf("failed create api handler: %+v", errors.WithStack(err))
	}
	defer apiHandler.Stop()
	http.Handle("/", apiHandler)

	log.Infof("server is listening at %s", addr)
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err == nil {
		log.Debugf("rlimit NOFILE: %d", rLimit.Cur)
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %+v", errors.WithStack(err))
	}
}
