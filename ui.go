package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func getfunc(w http.ResponseWriter, r *http.Request) {
	if n, e := strconv.Atoi(r.URL.Query().Get("n")); e != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, http.StatusText(http.StatusBadRequest))
	} else {
		w.Write(t.getFunc(n))
	}
}

func info(w http.ResponseWriter, r *http.Request) {
	bs := t.fi.Size()
	fs := float64(bs)
	ss := ""
	if bs > 1048576 {
		ss = fmt.Sprintf("%.2fmb", fs/1048576)
	} else if bs > 1024 {
		ss = fmt.Sprintf("%.2fkb", fs/1024)
	} else {
		ss = fmt.Sprintf("%db", bs)
	}
	w.Header().Add("Content-Type", "text/json")
	b, _ := json.Marshal(map[string]interface{}{
		"filename": t.fi.Name(),
		"filesize": ss,
	})
	w.Write(b)
}
