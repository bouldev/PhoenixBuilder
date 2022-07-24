package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

var storageDir = "./"

func main() {
	fmt.Println(filepath.Abs(storageDir))
	http.Handle("/", http.FileServer(http.Dir(storageDir)))
	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})
	port := "0.0.0.0:8083"
	fmt.Println("Server is running on port" + port)
	log.Fatal(http.ListenAndServe(port, nil))
	fmt.Println("Exit")
}
