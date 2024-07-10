package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"ga4ct/data"
	"ga4ct/templates"
)

func main() {
	eventStore, err := data.NewEventStore()
	if err != nil {
		log.Fatal(err)
	}
	defer eventStore.Close()

	// Handler function for storing events from pixel
	http.HandleFunc("/collect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		var event data.Event
		err = json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = eventStore.Insert(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("Inserted record with ID: %d\n", event.ID)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		templates.Placeholder().Render(r.Context(), w)
	})

	pixelFs := http.FileServer(http.Dir("./pixel"))
	http.Handle("/pixel/", http.StripPrefix("/pixel", pixelFs))

	assetsFs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets", assetsFs))

	fmt.Println("Server is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
