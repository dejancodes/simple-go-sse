package main

import (
	"fmt"
	"net/http"
)

var clients = make(map[string]http.ResponseWriter)

func main() {
	http.HandleFunc("/sse", sseHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/client", serveClient)
	http.HandleFunc("/", serveForm)
	http.ListenAndServe(":8080", nil)
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	secretKey := r.URL.Query().Get("secretKey")

	// Store the client's response writer for sending messages
	clients[secretKey] = w

	for {
		select {
		case <-r.Context().Done():
			delete(clients, secretKey)
			fmt.Println("Client closed connection.")
			return
		}
	}
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		message := r.FormValue("message")
		secretKey := r.FormValue("secretKey")

		if client, ok := clients[secretKey]; ok {
			fmt.Fprintf(client, "data: %s\n\n", message)
			flusher, ok := client.(http.Flusher)
			if !ok {
				http.Error(w, "Here not okey", http.StatusInternalServerError)
				return
			}
			flusher.Flush()
		} else {
			fmt.Fprintf(w, "Check your secret key.\n")
		}
	}
}

func serveForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "form.html")
}

func serveClient(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "client.html")
}
