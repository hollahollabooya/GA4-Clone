package main

import (
    "fmt"
    "net/http"
    "log"
    "encoding/json"
    "database/sql"
    "os"
    "strconv"
    "html/template"

    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
    "github.com/gorilla/websocket"
)

/* The key names here need to match the key names in the JSON object exactly
 * I don't know how to get around that but that's how we'll do it for now
 */
type Event struct {
    EventID    int
    EventName  string
    EventValue float64
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Event)

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

    go handleMessages()

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
        err, event.EventID = insertEvent(db, event)
        if err != nil {
            log.Fatal(err)
        }

        // Broadcast event to websockets
        broadcast <- event
    })

    http.HandleFunc("/ws", handleConnections)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        rows, err := db.Query(`SELECT * FROM events ORDER BY id DESC LIMIT 10`)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var events []Event
        for rows.Next() {
            var event Event
            if err := rows.Scan(&event.EventID, &event.EventName, &event.EventValue); err != nil {
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


    // fs := http.FileServer(http.Dir("../html"))
    // http.Handle("/", fs)

    pixelFs := http.FileServer(http.Dir("../pixel"))
    http.Handle("/pixel/", http.StripPrefix("/pixel", pixelFs))

    fmt.Println("Server is running on http://localhost:3000")
    log.Fatal(http.ListenAndServe(":3000", nil))
}

func enableCORS(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func insertEvent(db *sql.DB, event Event) (error, int) {
    sqlStatement := `
        INSERT INTO events (name, value)
        VALUES ($1, $2)
        RETURNING id`

    var eventID int
    err := db.QueryRow(sqlStatement, event.EventName, event.EventValue).Scan(&eventID)
    if err != nil {
        return err, 0
    }

    fmt.Printf("Inserted record with ID: %d\n", eventID)
    return nil, eventID
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    clients[ws] = true

    for {
        var event Event
        err := ws.ReadJSON(&event)
        if err != nil {
            delete(clients, ws)
            break
        }
    }
}

func handleMessages() {
    for {
        event := <-broadcast

        for client := range clients {
            err := client.WriteJSON(event)
            if err != nil {
                log.Printf("error: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}