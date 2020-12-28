package server

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strings"
)

var headerBlockList = map[string]bool{
	"report-to":      true,
	"content-length": true,
}

var defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:79.0) Gecko/20100101 Firefox/79.0"

var client = &http.Client{}

func getUri(uri string) (*http.Response, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", defaultUserAgent)

	return client.Do(req)
}

func stream(w http.ResponseWriter, r *http.Request, oriURL string, hostConfig *HostConfiguration) error {
	resp, err := getUri(oriURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Warning("failed to close stream with origin")
		}
	}()

	copyHTTPHeader(w, resp.StatusCode, resp.Header)

	contentType := resp.Header.Get("Content-Type")
	contentCategory := strings.ToLower(strings.Split(contentType, "/")[0])
	if contentCategory == "video" {
		streamBody(w, r, resp.Body, copyProxier())
	} else {
		streamBody(w, r, resp.Body, urlRewriteProxier(hostConfig))
	}

	return nil
}
func streamBody(w http.ResponseWriter, r *http.Request, resp io.ReadCloser, proxier func(w http.ResponseWriter, r *http.Request, resp io.ReadCloser), ) {
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			log.Infof("Stopped streaming uri %s", r.URL.Path)
			return
		default:
			proxier(w, r, resp)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return
		}
	}
}

func copyProxier() func(w http.ResponseWriter, r *http.Request, resp io.ReadCloser) {
	return func(w http.ResponseWriter, r *http.Request, resp io.ReadCloser) {
		_, _ = io.Copy(w, resp)
	}
}

func urlRewriteProxier(c *HostConfiguration) func(w http.ResponseWriter, r *http.Request, resp io.ReadCloser) {
	return func(w http.ResponseWriter, r *http.Request, resp io.ReadCloser) {
		name := getVarFromRequest("name", r)
		scanner := bufio.NewScanner(resp)
		var oldUrl string
		for scanner.Scan() {
			line := scanner.Text()
			if shouldProxy(line) {
				if !strings.HasPrefix(line, "http") {
					oldUrl = decodeUri(getVarFromRequest("uri", r))
					line = oldUrl[:strings.LastIndex(oldUrl, "/")+1] + line
				}
				_, _ = w.Write([]byte(encodeUri(name, line, c)))
			} else {
				_, _ = w.Write([]byte(line))
			}
			_, _ = w.Write([]byte("\n"))
		}
	}
}

func shouldProxy(line string) bool {
	return strings.HasPrefix(line, "http") || strings.HasSuffix(line, ".ts") || strings.HasSuffix(line, ".m3u") || strings.HasSuffix(line, ".m3u8")
}

func copyHTTPHeader(w http.ResponseWriter, statusCode int, header http.Header) {
	var ok bool
	for k, v := range header {
		if _, ok = headerBlockList[strings.ToLower(k)]; !ok {
			w.Header().Set(k, strings.Join(v, ", "))
		}
	}
	w.WriteHeader(statusCode)
}

func getVarFromRequest(v string, r *http.Request) string {
	vars := mux.Vars(r)

	return vars[v]
}

func encodeUri(name string, uri string, c *HostConfiguration) string {
	uriBase64 := base64.RawURLEncoding.EncodeToString([]byte(uri))

	return fmt.Sprintf(
		"http://%s:%d/%s/%s",
		c.Hostname,
		c.Port,
		name,
		uriBase64,
	)
}

func decodeUri(url string) string {
	uri, _ := base64.RawURLEncoding.DecodeString(url)
	return string(uri)
}
