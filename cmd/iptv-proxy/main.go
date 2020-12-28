package main

import (
	"flag"
	"fmt"
	"github.com/fugkco/iptv-proxy/pkg/server"
	"os"
	"strconv"
	"strings"
)

type m3uLists map[string]string

func (i *m3uLists) String() string {
	b := strings.Builder{}
	for k, v := range *i {
		b.WriteString(fmt.Sprintf("%s=%s,", k, strconv.Quote(v)))
	}
	return strings.TrimRight(b.String(), ",")
}

func (i *m3uLists) Set(value string) error {
	tokens := strings.SplitN(value, "=", 2)
	k := strings.TrimSpace(tokens[0])
	v := strings.TrimSpace(tokens[1])
	(*i)[k] = v
	return nil
}

func main() {
	var m3uPlaylists = m3uLists{}
	flag.Var(&m3uPlaylists, "playlists", "List of M3U files to proxy. Should be key pair values, e.g. --m3u bbc=https://bbc.co.uk/playlist.m3u. They key will be prefixed with all the URLs generates.")
	port := flag.Int64("port", 9090, "Port to expose the IPTVs endpoints")
	hostname := flag.String("hostname", "localhost", "Hostname or IP to expose the IPTVs endpoints")
	flag.Parse()

	if len(m3uPlaylists) == 0 {
		fmt.Printf("missing required -playlists flag\n")
		flag.Usage()
		os.Exit(1)
	}

	srv := server.New(*hostname, *port, m3uPlaylists)
	srv.Serve()
}
