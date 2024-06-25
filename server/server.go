package main

import (
    "fmt"
    "encoding/json"
    "net/http"
)

type Event struct {
    name  string    `json:"event_type"`
    value int       `json:"event_data"`
}

func main() {
    http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        var event Event
        if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        fmt.Sprintf("Event recieved. Event name: %s Event Value: %d \n", event.name, event.value)
    })

    fmt.Println("Server is running on http://localhost:3000")
    http.ListenAndServe(":3000", nil)
}