package server

import "net/http"

// HostConfiguration contains host infos
type HostConfiguration struct {
	Hostname string
	Port     int64
}

// Server Contain original m3u playlist and HostConfiguration
type Server struct {
	HostConfig *HostConfiguration
	Entries    map[string]string
	srv        *http.Server
}
