package main

import (
	"flag"
	"log"
	"net/http"
)

type ()

func main() {
	localPath := flag.String("path", "./store/", "local fs storage")
	addr := flag.String("addr", "127.0.0.1:8081", "address to listen")
	flag.Parse()

	fs, err := NewLocalFS(*localPath)
	if err != nil {
		panic(err)
	}
	ctl := Controller{
		FileSystem: fs,
	}
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
	}
	http.ListenAndServe(*addr, mw(ctl.Router()))

}
