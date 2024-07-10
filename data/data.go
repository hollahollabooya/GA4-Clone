package data

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

type EventStore struct {
	db *sql.DB
}

func NewEventStore() (*EventStore, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	host := os.Getenv("DB_HOST")
	portStr := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &EventStore{db: db}, nil
}

func (es *EventStore) Close() {
	es.db.Close()
}

func (es *EventStore) Insert(e *Event) error {
	// TODO: need to validate string length before inserting

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

	row := es.db.QueryRow(sqlStatement, e.AccountID, e.ClientID, e.SessionID,
		e.Name, e.Value, e.Timestamp, e.PageLocation, e.PageTitle, e.PageReferrer,
		e.UserAgent, e.ScreenResolution)

	return row.Scan(&e.ID)
}

type ModeledDimension struct {
	Label string
	sql   string
}

type ModeledMeasure struct {
	Label string
	sql   string
}

var (
	EventName = ModeledDimension{
		Label: "Event Name",
		sql:   `name`,
	}
	Date = ModeledDimension{
		Label: "Date",
		sql:   `TO_CHAR(timestamp, 'YYYY-MM-DD')`,
	}
)

var (
	EventCount = ModeledMeasure{
		Label: "Event Count",
		sql:   `COUNT(*)`,
	}
	EventValue = ModeledMeasure{
		Label: "Event Value",
		sql:   `ROUND(SUM(value),2)`,
	}
)

type Dimension string

type Measure float64

type Row struct {
	Dimensions []Dimension
	Measures   []Measure
}

type Table struct {
	DimensionHeaders []ModeledDimension
	MeasureHeaders   []ModeledMeasure
	Rows             []Row
}

func buildSQL(dimensions []ModeledDimension, measures []ModeledMeasure, limit int) string {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("SELECT ")

	if len(dimensions) > 0 && len(measures) == 0 {
		sqlBuilder.WriteString("DISTINCT ")
	}

	for i, dimension := range dimensions {
		sqlBuilder.WriteString(dimension.sql)
		if !(i == len(dimensions)-1 && len(measures) == 0) {
			sqlBuilder.WriteString(", ")
		}
	}

	for i, measure := range measures {
		sqlBuilder.WriteString(measure.sql)
		if i != len(measures)-1 {
			sqlBuilder.WriteString(", ")
		}
	}

	sqlBuilder.WriteString(" FROM events ")

	if len(dimensions) == 0 {
		sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d", limit))
		return sqlBuilder.String()
	}

	sqlBuilder.WriteString("GROUP BY ")

	for i := range dimensions {
		sqlBuilder.WriteString(strconv.Itoa(i + 1))
		if i != len(dimensions)-1 {
			sqlBuilder.WriteString(", ")
		}
	}

	sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	return sqlBuilder.String()
}

func Retrieve(db *sql.DB, modeledDimensions []ModeledDimension, modeledMeasures []ModeledMeasure) (*Table, error) {
	if len(modeledDimensions) == 0 && len(modeledMeasures) == 0 {
		return nil, sql.ErrNoRows
	}

	rows, err := db.Query(buildSQL(modeledDimensions, modeledMeasures, 10))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	table := Table{
		DimensionHeaders: modeledDimensions,
		MeasureHeaders:   modeledMeasures,
	}
	for rows.Next() {
		row := Row{
			Dimensions: make([]Dimension, len(modeledDimensions)),
			Measures:   make([]Measure, len(modeledMeasures)),
		}

		scanVals := make([]any, len(modeledDimensions)+len(modeledMeasures))
		for i := range row.Dimensions {
			scanVals[i] = &row.Dimensions[i]
		}
		for i := range row.Measures {
			scanVals[len(modeledDimensions)+i] = &row.Measures[i]
		}

		if err := rows.Scan(scanVals...); err != nil {
			return nil, err
		}
		table.Rows = append(table.Rows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &table, nil
}
