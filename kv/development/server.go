package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/grafana/dskit/dns"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/dskit/kv/memberlist"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	defaultLocalAddress   = "127.0.0.1"
	defaultMemberlistPort = 7946
	defaultServerPort     = 8080
)

type config struct {
	BindAddr string
	BindPort int
}

type server struct {
	cfg        config
	kv         *memberlist.KV
	httpServer *http.Server
	logger     *log.Logger
}

func newServer(cfg config) *server {
	router := http.NewServeMux()
	router.HandleFunc("GET /", defaultHandler)            // show entire key-value store
	router.HandleFunc("GET /{key}", getItemHandler)       // get value for key
	router.HandleFunc("POST /{key}", postItemHandler)     // create new key-value pair
	router.HandleFunc("PUT /{key}", putItemHandler)       // update value for key
	router.HandleFunc("DELETE /{key}", deleteItemHandler) // delete key-value pair

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.BindPort),
		Handler: router,
	}

	logger := log.With(log.NewLogfmtLogger(os.Stdout), level.AllowDebug())

	var mbConfig memberlist.KVConfig
	flagext.DefaultValues(&cfg)

	mbConfig.TCPTransport = memberlist.TCPTransportConfig{
		BindAddrs: []string{bindAddr},
		BindPort:  bindPort,
	}
	// joinmembers are the addresses of peers who are already in the memberlist group.
	// Usually provided if this peer is trying to join an existing cluster.
	// Generally you start the very first peer without `joinmembers`, but start all
	// other peers with at least one `joinmembers`.
	if len(joinmembers) > 0 {
		cfg.JoinMembers = joinmembers
	}

	// resolver defines how each peers IP address should be resolved.
	// We use default resolver comes with Go.
	resolver := dns.NewProvider(log.With(logger, "component", "dns"), prometheus.NewPedanticRegistry(), dns.GolangResolverType)

	mbConfig.NodeName = bindaddr
	mbConfig.StreamTimeout = 5 * time.Second

	return &server{
		cfg:        cfg,
		kv:         kv,
		httpServer: srv,
		logger:     logger,
	}
}

func main() {
	cfg := config{
		BindAddr: defaultLocalAddress,
		BindPort: defaultServerPort,
	}

	srv := newServer(cfg)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	<-stopChan

	if err := srv.httpServer.Shutdown(context.Background()); err != nil {
		log.Fatalf("server failed to shutdown: %v", err)
	}
}

func deleteItemHandler(writer http.ResponseWriter, request *http.Request) {

}

func putItemHandler(writer http.ResponseWriter, request *http.Request) {

}

func postItemHandler(writer http.ResponseWriter, request *http.Request) {

}

func getItemHandler(writer http.ResponseWriter, request *http.Request) {

}

func defaultHandler(writer http.ResponseWriter, request *http.Request) {

}
