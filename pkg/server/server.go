/*
 * Iptv-Proxy is a project to proxify an m3u file.
 * Copyright (C) 2020  Pierre-Emmanuel Jacquier
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package server

import (
	"context"
	"fmt"
	"github.com/google/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type key int
const (
	requestIDKey key = 0
)

var log = logger.Init("server", false, false, os.Stdout)
var httpLog = logger.Init("http", false, false, os.Stdout)

// New initialize a new server configuration
func New(hostname string, port int64, entries map[string]string) *Server {
	parsedUrlsMap := make(map[string]string, len(entries))
	for name, urlStr := range entries {
		uri, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalf("[iptv-proxy] ERROR invalid URL for %s (%s), got error: %s", name, urlStr, err)
		}
		parsedUrlsMap[name] = uri.String()
	}

	return &Server{
		HostConfig: &HostConfiguration{
			Hostname: hostname,
			Port:     port,
		},
		Entries: parsedUrlsMap,
	}
}

// Serve the iptv-proxy api
func (s *Server) Serve() {
	r := mux.NewRouter()

	group := r.PathPrefix("/").Subrouter()
	s.routes(group)

	nextRequestID := func() string {
		return uuid.New().String()
	}

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.HostConfig.Port),
		Handler:      tracing(nextRequestID)(logging(r)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			log.Fatalf("failed to serve iptv-proxy due to err %s", err)
		}
	}()

	log.Infof("started on %s:%d", s.HostConfig.Hostname, s.HostConfig.Port)

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		httpLog.Fatalf("failed to shutdown server due to error %s", err)
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			requestID, ok := r.Context().Value(requestIDKey).(string)
			if !ok {
				requestID = "unknown"
			}

			username := "-"
			if r.URL.User != nil {
				if name := r.URL.User.Username(); name != "" {
					username = name
				}
			}

			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}

			uri := r.RequestURI
			// Requests using the CONNECT method over HTTP/2.0 must use
			// the authority field (aka r.Host) to identify the target.
			// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
			if r.ProtoMajor == 2 && r.Method == "CONNECT" {
				uri = r.Host
			}
			if uri == "" {
				uri = r.URL.RequestURI()
			}

			// todo add statuscode + size, see https://godoc.org/github.com/gorilla/handlers#CombinedLoggingHandler
			httpLog.Infof(`%s - %s [%s] "%s %s %s"`, host, username, requestID, r.Method, uri, r.Proto)
		}()
		next.ServeHTTP(w, r)
	})
}
