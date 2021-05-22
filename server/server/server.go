package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.etcd.io/etcd/clientv3"
)

var DEFAULT_ENDPOINTS []string = []string{"127.0.0.1:2379", "127.0.0.1:22379", "127.0.0.1:32379"}
var DEFAULT_DIAL_TIMEOUT time.Duration = 3 * time.Second

type Server struct {
	Router     *httprouter.Router
	etcdClient *clientv3.Client

	debugMode bool
}

func (s *Server) Teardown() {
	// From the client/v3 docs:
	// "Make sure to close the client after using it. If the client is not
	// closed, the connection will have leaky goroutines."
	s.etcdClient.Close()

	fmt.Println("RBDNS: Tearing down server. Goodbye!")
}

func (s *Server) HelloWorldGet(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {

	if s.debugMode {
		fmt.Println("helloWorld: in")
	}

	fmt.Fprintln(w, "Hello world!")
}

func (s *Server) AddRecordGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if s.debugMode {
		fmt.Printf("addRecord(%s, %s): in\n", key, value)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DIAL_TIMEOUT)
	_, err := s.etcdClient.Put(ctx, key, value)
	cancel()

	if err != nil {
		fmt.Printf("addRecord(%s, %s): in\n", key, value)
		fmt.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Internal server error")

		return
	}

	if s.debugMode {
		fmt.Printf("addRecord(%s, %s): success\n", key, value)
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

func (s *Server) QueryGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	key := r.URL.Query().Get("key")

	if s.debugMode {
		fmt.Printf("query(%s): in\n", key)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DIAL_TIMEOUT)
	resp, err := s.etcdClient.Get(ctx, key)
	cancel()

	if err != nil {
		fmt.Printf("query(%s): error\n", key)
		fmt.Println(err)

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if s.debugMode {
		fmt.Printf("query(%s): success. Value: %s\n", key, resp.Kvs[0].Value)
	}

	if len(resp.Kvs) == 0 {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "null")
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, string(resp.Kvs[0].Value))
	}
}

func New(debugMode bool) Server {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   DEFAULT_ENDPOINTS,
		DialTimeout: DEFAULT_DIAL_TIMEOUT,
	})
	if err != nil {
		fmt.Println("Failed to initialize etcd client")
		fmt.Println("Error:")
		fmt.Println(err)
		os.Exit(1)
	}

	router := httprouter.New()
	s := Server{
		Router:     router,
		etcdClient: client,
		debugMode:  debugMode,
	}

	router.GET("/helloWorld", s.HelloWorldGet)
	router.GET("/addRecord", s.AddRecordGet)
	router.GET("/query", s.QueryGet)

	return s
}
