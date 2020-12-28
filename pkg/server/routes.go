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
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) routes(r *mux.Router) {
	r.HandleFunc("/{name:[a-z]+}.m3u", s.getPlaylist()).Methods("GET")
	// todo this is dangerous, this essentially proxies _anything_ as long as it can be decoded.
	//  we should whitelist the specific URLs that we're willing to stream as we proxy them
	r.HandleFunc("/{name:[a-z]+}/{uri:[a-zA-Z0-9\\-_]+}", s.reverseProxy()).Methods("GET") // should match base64.encodeURL
}

func (s *Server) getPlaylist() func(w http.ResponseWriter, r *http.Request) {
	for name := range s.Entries {
		log.Infof("registering http://%s:%d/%s.m3u", s.HostConfig.Hostname, s.HostConfig.Port, name)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		name := getVarFromRequest("name", r)

		uri, ok := s.Entries[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("stream group %s does not exist", name)))
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, name+".m3u"))
		w.Header().Set("Content-Type", "application/octet-stream")

		err := stream(w, r, uri, s.HostConfig)
		if err != nil {
			log.Warningf("failed to abort context, error was: %s", err.Error())
		}
	}
}

func (s *Server) reverseProxy() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		name := vars["name"]
		_, ok := s.Entries[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("stream group %s does not exist", name)))
			return
		}

		uri := decodeUri(vars["uri"])
		err := stream(w, r, uri, s.HostConfig)
		if err != nil {
			log.Warningf("failed to abort context, error was: %s", err.Error())
		}
	}
}
