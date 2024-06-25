package main

import (
    "fmt"
    "net/http"
    "log"
    "io/ioutil"
)

type Event struct {
    name  string    `json:"event_name"`
    value int       `json:"event_value"`
}

func main() {
    http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
        enableCORS(&w)

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
        if r.Method != http.MethodPost {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        // Read the request body
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Error reading request body: "+err.Error(), http.StatusBadRequest)
            return
        }

        // Print the request body for debugging
        fmt.Println("Request Body:", string(body))


        // I can read the body, but I can't properly decode the JSON into the struct woot woot
    })

    fmt.Println("Server is running on http://localhost:3000")
    log.Fatal(http.ListenAndServe(":3000", nil))
}

func enableCORS(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}