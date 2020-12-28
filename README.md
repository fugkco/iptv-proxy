# Iptv Proxy

[![Actions Status](https://github.com/fugkco/iptv-proxy/workflows/CI/badge.svg)](https://github.com/fugkco/iptv-proxy/actions?query=workflow%3ACI)

## Description

Iptv-proxy is a project to proxy a m3u file, and all its tracks.

### M3U

M3U service convert an iptv m3u file into a web proxy server.

It points all the original tracks to the proxied tracks. The code converts _all_ lines in playlist that are URLs. It also attempts to convert any relative paths in a playlist.

## WARNING: This can proxy any URL

This implementation iptv-proxy can proxy literally any URL due to how it was implemented. This is insecure. The URLs are simply base64 encoded when written to the playlist, and then later decoded and proxied. This seemed the simplest mechanism due to some playlists being nested multiple levels.

You should aim to keep this as closed off as possible, if you can, only run on localhost.

### M3u Example

Original iptv m3u file

```m3u
#EXTM3U
#EXTINF:-1 tvg-ID="examplechanel1.com" tvg-name="chanel1" tvg-logo="http://ch.xyz/logo1.png" group-title="USA HD",CHANEL1-HD
http://iptvexample.net:1234/12/test/1
#EXTINF:-1 tvg-ID="examplechanel2.com" tvg-name="chanel2" tvg-logo="http://ch.xyz/logo2.png" group-title="USA HD",CHANEL2-HD
http://iptvexample.net:1234/13/test/2
#EXTINF:-1 tvg-ID="examplechanel3.com" tvg-name="chanel3" tvg-logo="http://ch.xyz/logo3.png" group-title="USA HD",CHANEL3-HD
http://iptvexample.net:1234/14/test/3
#EXTINF:-1 tvg-ID="examplechanel4.com" tvg-name="chanel4" tvg-logo="http://ch.xyz/logo4.png" group-title="USA HD",CHANEL4-HD
http://iptvexample.net:1234/15/test/4
```

What M3U proxy IPTV do
 - convert channels url to new endpoints
 - convert original m3u file with new routes pointing to the proxy

Start proxy server example

```Bash
iptv-proxy --playlists mylist=http://example.com/get.php?username=user&password=pass&type=m3u_plus&output=m3u8 \
             --port 8080 \
             --hostname proxyexample.com
```


This gives you a new m3u playlist on the endpoint `/mylist.m3u` in our example

```
http://proxyexample.com:8080/mylist.m3u
```

All the new routes pointing on your proxy server
```m3u
#EXTM3U
#EXTINF:-1 tvg-ID="examplechanel1.com" tvg-name="chanel1" tvg-logo="http://ch.xyz/logo1.png" group-title="USA HD",CHANEL1-HD
http://proxyexample.com:8080/mylist/12/test/1?username=test&password=passwordtest
#EXTINF:-1 tvg-ID="examplechanel2.com" tvg-name="chanel2" tvg-logo="http://ch.xyz/logo2.png" group-title="USA HD",CHANEL2-HD
http://proxyexample.com:8080/mylist/13/test/2?username=test&password=passwordtest
#EXTINF:-1 tvg-ID="examplechanel3.com" tvg-name="chanel3" tvg-logo="http://ch.xyz/logo3.png" group-title="USA HD",CHANEL3-HD
http://proxyexample.com:8080/mylist/14/test/3?username=test&password=passwordtest
#EXTINF:-1 tvg-ID="examplechanel4.com" tvg-name="chanel4" tvg-logo="http://ch.xyz/logo4.png" group-title="USA HD",CHANEL4-HD
http://proxyexample.com:8080/mylist/15/test/4?username=test&password=passwordtest
```

## Installation

Download lasted [release](https://github.com/fugkco/iptv-proxy/releases)

Or

`% go install` in root repository

## With Docker

### Prerequisite

- Add an m3u URL in `docker-compose.yml` or add local file in `iptv` folder
- Expose same container port as the `--port` flag 

### Start

`docker-compose` sample:
```Yaml
version: "3"
services:
  iptv-proxy:
    build:
      context: .
      dockerfile: Dockerfile
    volumes:
      # If your are using local m3u file instead of m3u remote file
      # put your m3u file in this folder
      - ./iptv:/root/iptv
    container_name: "iptv-proxy"
    restart: on-failure
    commad: --playlists mylist=http://proxyexample.com:8080/mylist.m3u
    expose:
      - 8080
```

Addition arguments you can add to the `command` section:
```
Usage of iptv-proxy:
  -hostname string
        Hostname or IP to expose the IPTVs endpoints (default "localhost")
  -playlists value
        List of M3U files to proxy. Should be key pair values, e.g. --m3u bbc=http://example.com/playlist.m3u. They key will be prefixed with all the URLs generates.
  -port int
        Port to expose the IPTVs endpoints (default 8080)
```

Then you start it
```
% docker-compose up -d
```

### Thanks
This is heavily modified code forked from [pierre-emmanuelJ/iptv-proxy](https://github.com/pierre-emmanuelJ/iptv-proxy).
