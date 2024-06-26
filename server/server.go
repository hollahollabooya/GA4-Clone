package main

import (
    "fmt"
    "net/http"
    "log"
    "encoding/json"
)

/* The keys here need to match the keys in the JSON object exactly
 * I don't know how to get around that but that's how we'll do it for now
 */
type Event struct {
    EventName  string
    EventValue int
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
        var event Event
        err := json.NewDecoder(r.Body).Decode(&event)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        fmt.Printf("Event: %+v \n", event)
    })

    fmt.Println("Server is running on http://localhost:3000")
    log.Fatal(http.ListenAndServe(":3000", nil))
}

func enableCORS(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}