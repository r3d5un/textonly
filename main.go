package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Hello from textonly"))
}

func readPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Display a specific post with ID %d...", id)
}

func posts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from posts"))
}

func about(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from about"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/post/read/", readPost)
	mux.HandleFunc("/post/", posts)
	mux.HandleFunc("/about/", posts)

	log.Println("Starting server on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
