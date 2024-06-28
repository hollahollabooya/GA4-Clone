package main

import (
    "fmt"
    "net/http"
    "log"
    "encoding/json"
    "database/sql"
    "os"
    "strconv"

    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
)

/* The key names here need to match the key names in the JSON object exactly
 * I don't know how to get around that but that's how we'll do it for now
 */
type Event struct {
    EventName  string
    EventValue float64
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
        err = json.NewDecoder(r.Body).Decode(&event)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        fmt.Printf("Event: %+v \n", event)

        // Insert the event into the DB
        err = insertEvent(db, event)
        if err != nil {
            log.Fatal(err)
        }
    })

    fmt.Println("Server is running on http://localhost:3000")
    log.Fatal(http.ListenAndServe(":3000", nil))
}

func enableCORS(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "https://hollahollabooya.github.io")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func insertEvent(db *sql.DB, event Event) error {
    sqlStatement := `
        INSERT INTO events (name, value)
        VALUES ($1, $2)
        RETURNING id`

    var eventID int
    err := db.QueryRow(sqlStatement, event.EventName, event.EventValue).Scan(&eventID)
    if err != nil {
        return err
    }

    fmt.Printf("Inserted record with ID: %d\n", eventID)
    return nil
}