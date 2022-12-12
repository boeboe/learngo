package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

func newMux(todoFile string) http.Handler {
	m := http.NewServeMux()
	mu := &sync.Mutex{}

	m.HandleFunc("/", rootHandler)
	t := todoRouter(todoFile, mu)

	m.Handle("/todo", http.StripPrefix("/todo", t))
	m.Handle("/todo/", http.StripPrefix("/todo/", t))
	return m
}

func replyTextContent(w http.ResponseWriter, req *http.Request, status int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content))
}

func replyJSONContent(w http.ResponseWriter, req *http.Request, status int, resp *todoResponse) {
	body, err := json.Marshal(resp)
	if err != nil {
		replyError(w, req, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}

func replyError(w http.ResponseWriter, req *http.Request, status int, message string) {
	log.Printf("%s %s: Error: %d %s", req.URL, req.Method, status, message)
	http.Error(w, http.StatusText(status), status)
}
