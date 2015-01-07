package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
)

var filename string
var listen = "127.0.0.1:8888"

var t = &trace{}

func serve() {
	http.HandleFunc("/api/info.json", info)
	http.HandleFunc("/api/func.json", getfunc)
	if webRoot, err := filepath.Abs("./public_html"); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Serving %s", webRoot)
		http.Handle("/", http.FileServer(http.Dir(webRoot)))
	}
	log.Fatal(http.ListenAndServe(listen, nil))
}

func init() {
	flag.StringVar(&filename, "f", filename, "Trace file to explore")
	flag.StringVar(&listen, "l", listen, "address:port to listen on for serving the HTTP interface")
}

func main() {
	flag.Parse()
	if fi, err := os.Lstat(filename); err != nil {
		log.Fatal(err)
	} else {
		t.fi = fi
	}
	if fp, err := os.Open(filename); err != nil {
		log.Fatal(err)
	} else {
		t.fp = fp
	}

	fmt.Printf("Building function index ")
	s := spinner.New(spinner.CharSets[11], 75*time.Millisecond)
	s.Start()
	start := time.Now()
	t.index()
	s.Stop()
	fmt.Println(" done in", time.Now().Sub(start).Seconds(), "seconds")
	serve()
}
