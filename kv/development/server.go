package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grafana/dskit/dns"
	"github.com/grafana/dskit/flagext"
	"github.com/grafana/dskit/kv/memberlist"
	"github.com/grafana/dskit/services"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	defaultLocalAddress   = "127.0.0.1"
	defaultMemberlistPort = 7946
	defaultServerPort     = 8100
)

func main() {
	var bindaddr string
	var bindport int
	var joinmember string

	flag.StringVar(&bindaddr, "bindaddr", defaultLocalAddress, "bindaddr for this specific peer")
	flag.IntVar(&bindport, "bindport", defaultMemberlistPort, "bindport for this specific peer")
	flag.StringVar(&joinmember, "join-member", "", "peer addr that is part of existing cluster")

	flag.Parse()

	ctx := context.Background()
	logger := log.With(log.NewLogfmtLogger(os.Stdout), level.AllowDebug())

	joinmembers := make([]string, 0)
	if joinmember != "" {
		joinmembers = append(joinmembers, joinmember)
	}

	var mbConfig memberlist.KVConfig
	flagext.DefaultValues(&mbConfig)

	// non-default options
	mbConfig.RandomizeNodeName = false
	mbConfig.ClusterLabel = "cluster"
	mbConfig.MessageHistoryBufferBytes = 1024 * 1024 * 10

	mbConfig.TCPTransport = memberlist.TCPTransportConfig{
		BindAddrs:      []string{bindaddr},
		BindPort:       bindport,
		TransportDebug: true,
	}
	// joinmembers are the addresses of peers who are already in the memberlist group.
	// Usually provided if this peer is trying to join an existing cluster.
	// Generally you start the very first peer without `joinmembers`, but start all
	// other peers with at least one `joinmembers`.
	if len(joinmember) > 0 {
		mbConfig.JoinMembers = joinmembers
	}

	// resolver defines how each peers IP address should be resolved.
	// We use default resolver comes with Go.
	resolver := dns.NewProvider(log.With(logger, "component", "dns"), prometheus.NewPedanticRegistry(), dns.GolangResolverType)

	mbConfig.NodeName = bindaddr
	mbConfig.StreamTimeout = 5 * time.Second

	kvSvc := memberlist.NewKVInitService(&mbConfig, log.With(logger, "component", "memberlist"), resolver, prometheus.NewPedanticRegistry())
	if err := services.StartAndAwaitRunning(ctx, kvSvc); err != nil {
		panic(err)
	}
	defer services.StopAndAwaitTerminated(ctx, kvSvc)

	_, err := kvSvc.GetMemberlistKV()
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(bindaddr, strconv.Itoa(defaultServerPort)))
	if err != nil {
		panic(err)
	}

	fmt.Println("listening on ", listener.Addr())

	router := http.NewServeMux()
	router.Handle("GET /", kvSvc)                         // show entire key-value store
	router.HandleFunc("GET /{key}", getItemHandler)       // get value for key
	router.HandleFunc("POST /{key}", postItemHandler)     // create new key-value pair
	router.HandleFunc("PUT /{key}", putItemHandler)       // update value for key
	router.HandleFunc("DELETE /{key}", deleteItemHandler) // delete key-value pair

	panic(http.Serve(listener, router))

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
