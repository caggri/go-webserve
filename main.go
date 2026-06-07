package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.status == 0 {
		lrw.status = http.StatusOK
	}

	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytes += n
	return n, err
}

func folderHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	lrw := &loggingResponseWriter{ResponseWriter: w}

	path := "." + r.URL.Path
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "."
	}

	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("ERROR stat %q: %v", path, err)
		}
	} else if !info.IsDir() {
		fileName := info.Name()
		lrw.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	}

	http.FileServer(http.Dir("./")).ServeHTTP(lrw, r)
	log.Printf("%s %s %s %d %dB %s",
		r.RemoteAddr,
		r.Method,
		r.URL.Path,
		lrw.status,
		lrw.bytes,
		time.Since(start).Truncate(time.Millisecond),
	)
}

func main() {
	var port string
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) == 0 {
		port = "8888"
	} else {
		port = argsWithoutProg[0]
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("starting simple HTTP file server on :%s", port)

	http.HandleFunc("/", folderHandler)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
