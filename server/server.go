package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/* The key names here need to match the key names in the JSON object exactly
 * I don't know how to get around that but that's how we'll do it for now
 */
type Event struct {
	ID               int       `json:"-"`
	AccountID        string    `json:"account_id"`
	ClientID         string    `json:"client_id"`
	SessionID        string    `json:"session_id"`
	Name             string    `json:"event_name"`
	Value            float64   `json:"event_value"`
	Timestamp        time.Time `json:"timestamp"`
	PageLocation     string    `json:"page_location"`
	PageTitle        string    `json:"page_title"`
	PageReferrer     string    `json:"page_referrer"`
	UserAgent        string    `json:"user_agent"`
	ScreenResolution string    `json:"screen_resolution"`
}

func main() {
	// Load the environment credentials
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read database credentials from environment variables
	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port number: %v", portStr)
	}

	// Setup database connection
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the database")

	// Handler function for storing events from pixel
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		var event Event
		err = json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// I need to insert some logic to truncate the string fields here so things don't get too brazy

		// fmt.Printf("IP: %+v\n", r.RemoteAddr)
		fmt.Printf("Event: %+v \n", event)

		// Insert the event into the DB
		sqlStatement := `
        INSERT INTO events (
			account_id,
			client_id,
			session_id,
			name, 
			value, 
			timestamp,
			page_location,
			page_title,
			page_referrer,
			user_agent,
			screen_resolution
		)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id`

		var eventID int
		err := db.QueryRow(sqlStatement, event.AccountID, event.ClientID, event.SessionID,
			event.Name, event.Value, event.Timestamp, event.PageLocation, event.PageTitle,
			event.PageReferrer, event.UserAgent, event.ScreenResolution).Scan(&eventID)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Inserted record with ID: %d\n", eventID)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sqlStatement := `
        SELECT
			id,
			name,
			value 
		FROM events
		ORDER BY id DESC
		LIMIT 10`

		rows, err := db.Query(sqlStatement)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var events []Event
		for rows.Next() {
			var event Event
			if err := rows.Scan(&event.ID, &event.Name, &event.Value); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			events = append(events, event)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles("../html/index.html"))
		tmpl.Execute(w, map[string]interface{}{"events": events})
	})

	pixelFs := http.FileServer(http.Dir("../pixel"))
	http.Handle("/pixel/", http.StripPrefix("/pixel", pixelFs))

	fmt.Println("Server is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
