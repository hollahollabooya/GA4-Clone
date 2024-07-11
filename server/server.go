package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"ga4ct/data"
	"ga4ct/templates"
)

func main() {
	eventStore, err := data.NewEventStore()
	if err != nil {
		log.Fatal(err)
	}
	defer eventStore.Close()

	// Handler function for storing events sent from pixels
	http.HandleFunc("/collect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

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

		res, err := eventStore.NewQuery().Dimensions(data.EventName).
			Measures(data.EventCount, data.EventValue).Limit(10).Query()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		table, err := res.Table()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		templates.Index(table).Render(r.Context(), w)
	})

	pixelFs := http.FileServer(http.Dir("./pixel"))
	http.Handle("/pixel/", http.StripPrefix("/pixel", pixelFs))

	assetsFs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets", assetsFs))

	fmt.Println("Server is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
