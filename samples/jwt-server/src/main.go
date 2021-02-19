// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	// nolint: lll
	jwtKey = "{ \"keys\":[ {\"e\":\"AQAB\",\"kid\":\"tT_w9LRNrY7wJalGsTYSt7rutZi86Gvyc0EKR4CaQAw\",\"kty\":\"RSA\",\"n\":\"raJ7ZEhMfrBUo2werGKOow9an1B6Ukc6dKY2hNi10eaQe9ehJCjLpmJpePxoqaCi2VYt6gncLfhEV71JDGsodbfYMlaxwWTt6lXBcjlVXHWDXLC45rHVfi9FjSSXloHqmSStpjv3mrW3R6fx2VeVVP_mrA6ZHtcynq6ecJqO11STvVoeeM3lEsASVSWsUrKltC1Crfo0sI7YG34QjophVTEi8B9gVepAJZV-Bso5sinRABnxfLUM7DU5c8MO114uvXThgSIuAOM9PbViSC3X6Y9Gsjsy881HGO-EJaUCrwSWnwQW5sp0TktrYL70-M4_ug-X51Yt_PErmncKupx8Hw\"}]}"
)

var httpPort = flag.String("http", "8000", "HTTP server port")

// JWTServer implements the sample server that serves jwt keys.
type JWTServer struct {
	httpServer *http.Server
	// For test only
	httpPort chan int
}

// ServeHTTP serves the JWT Keys.
func (s *JWTServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(string(jwtKey)))
}

func (s *JWTServer) startHTTP(address string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		log.Printf("Stopped JWT HTTP server")
	}()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}
	// Store the port for test only.
	s.httpPort <- listener.Addr().(*net.TCPAddr).Port
	s.httpServer = &http.Server{Handler: s}

	log.Printf("Starting HTTP server at %s", listener.Addr())
	if err := s.httpServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func (s *JWTServer) run(httpAddr string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go s.startHTTP(httpAddr, &wg)
	wg.Wait()
}

func (s *JWTServer) stop() {
	s.httpServer.Close()
}

func NewJwtServer() *JWTServer {
	return &JWTServer{
		httpPort: make(chan int, 1),
	}
}

func main() {
	flag.Parse()
	s := NewJwtServer()
	go s.run(fmt.Sprintf(":%s", *httpPort))
	defer s.stop()

	// Wait for the process to be shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
