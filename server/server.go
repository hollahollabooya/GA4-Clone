package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"ga4ct/data"
	"ga4ct/templates"
)

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
		var event data.Event
		err = json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// I need to insert some logic to truncate the string fields here so things don't get too brazy

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

		// Get the aggregated data by date
		dimensions := []data.ModeledDimension{data.Date}
		measures := []data.ModeledMeasure{data.EventCount}

		lineChartData, err := data.Retrieve(db, dimensions, measures)
		if err != nil {
			log.Fatal(err)
		}

		// Get the aggregated data by event name
		dimensions = []data.ModeledDimension{data.EventName}
		measures = []data.ModeledMeasure{data.EventCount}

		barChartData, err := data.Retrieve(db, dimensions, measures)
		if err != nil {
			log.Fatal(err)
		}

		templates.Index(lineChartData, barChartData).Render(r.Context(), w)
	})

	pixelFs := http.FileServer(http.Dir("./pixel"))
	http.Handle("/pixel/", http.StripPrefix("/pixel", pixelFs))

	assetsFs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets", assetsFs))

	jsFs := http.FileServer(http.Dir("./node_modules"))
	http.Handle("/node_modules/", http.StripPrefix("/node_modules", jsFs))

	fmt.Println("Server is running on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
