package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		fmt.Fprintf(w, "Hello World!")
		fmt.Println(t.Format(time.RFC822Z), ":A new request is received")
	})

	t := time.Now()
	fmt.Println(t.Format(time.RFC822Z), ":Server is running") 

	http.ListenAndServe(":80", nil)
}
