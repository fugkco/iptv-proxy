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
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var log = logger.Init("server", false, false, os.Stdout)

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

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.HostConfig.Hostname, s.HostConfig.Port),
		Handler:      r,
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
		log.Fatalf("failed to shutdown server due to error %s", err)
	}
}
