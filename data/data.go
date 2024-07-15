package data

import (
	"database/sql"
	"errors"
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
	// I might need to not do this work here if I need a Session Store too?
	// Something to revisit
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
	// Database table has limits on string length for some column
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

type modeledDimension struct {
	label string
	sql   string
}

func (md *modeledDimension) Label() string {
	return md.label
}

type modeledMeasure struct {
	label string
	sql   string
}

func (mm *modeledMeasure) Label() string {
	return mm.label
}

var (
	EventName = modeledDimension{
		label: "Event Name",
		sql:   `name`,
	}
	Date = modeledDimension{
		label: "Date",
		sql:   `TO_CHAR(timestamp, 'YYYY-MM-DD')`,
	}
)

var (
	EventCount = modeledMeasure{
		label: "Event Count",
		sql:   `COUNT(*)`,
	}
	EventValue = modeledMeasure{
		label: "Event Value",
		sql:   `ROUND(SUM(value),2)`,
	}
)

type query struct {
	db         *sql.DB
	dimensions []modeledDimension
	measures   []modeledMeasure
	limit      uint32 // Limit of 0 drops LIMIT clause from SQL statement
}

type result struct {
	query *query
	rows  *sql.Rows
}

func (es *EventStore) NewQuery() *query {
	return &query{db: es.db}
}

func (q *query) Dimensions(mds ...modeledDimension) *query {
	q.dimensions = mds
	return q
}

func (q *query) Measures(mms ...modeledMeasure) *query {
	q.measures = mms
	return q
}

func (q *query) Limit(limit uint32) *query {
	q.limit = limit
	return q
}

func (q *query) buildSQL() string {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("SELECT ")

	if len(q.dimensions) > 0 && len(q.measures) == 0 {
		sqlBuilder.WriteString("DISTINCT ")
	}

	for i, dimension := range q.dimensions {
		sqlBuilder.WriteString(dimension.sql)
		if !(i == len(q.dimensions)-1 && len(q.measures) == 0) {
			sqlBuilder.WriteString(", ")
		}
	}

	for i, measure := range q.measures {
		sqlBuilder.WriteString(measure.sql)
		if i != len(q.measures)-1 {
			sqlBuilder.WriteString(", ")
		}
	}

	sqlBuilder.WriteString(" FROM events")

	// To-Do: revisit this and debug
	if len(q.measures) == 0 && q.limit != 0 {
		sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d", q.limit))
		return sqlBuilder.String()
	}

	if len(q.dimensions) == 0 {
		return sqlBuilder.String()
	}

	sqlBuilder.WriteString(" GROUP BY ")

	for i := range q.dimensions {
		sqlBuilder.WriteString(strconv.Itoa(i + 1))
		if i != len(q.dimensions)-1 {
			sqlBuilder.WriteString(", ")
		}
	}

	if q.limit != 0 {
		sqlBuilder.WriteString(fmt.Sprintf(" LIMIT %d", q.limit))
	}

	return sqlBuilder.String()
}

func (q *query) Query() (*result, error) {
	if len(q.dimensions) == 0 && len(q.measures) == 0 {
		return nil, sql.ErrNoRows
	}

	rows, err := q.db.Query(q.buildSQL())
	if err != nil {
		return nil, err
	}

	return &result{rows: rows, query: q}, nil
}

func (r *result) Close() {
	r.rows.Close()
}

type Row struct {
	Dimensions []string
	Measures   []float64
}

type Table struct {
	DimensionHeaders []modeledDimension
	MeasureHeaders   []modeledMeasure
	Rows             []Row
}

func (r *result) Table() (*Table, error) {
	table := Table{
		DimensionHeaders: r.query.dimensions,
		MeasureHeaders:   r.query.measures,
	}

	for r.rows.Next() {
		row := Row{
			Dimensions: make([]string, len(table.DimensionHeaders)),
			Measures:   make([]float64, len(table.MeasureHeaders)),
		}

		scanVals := make([]any, len(table.DimensionHeaders)+len(table.MeasureHeaders))
		for i := range row.Dimensions {
			scanVals[i] = &row.Dimensions[i]
		}
		for i := range row.Measures {
			scanVals[len(table.DimensionHeaders)+i] = &row.Measures[i]
		}

		if err := r.rows.Scan(scanVals...); err != nil {
			return nil, err
		}
		table.Rows = append(table.Rows, row)
	}
	if err := r.rows.Err(); err != nil {
		return nil, err
	}

	return &table, nil
}

// These data structures are defined to play nice with charts.js
// https://www.chartjs.org/docs/latest/general/data-structures.html#object
type dataPoint struct {
	X string  `json:"x"`
	Y float64 `json:"y"`
}

type DataSet struct {
	Data []dataPoint `json:"data"`
}

var ErrResultMalformed = errors.New("result malformed for this data transformation")

/* I might be able to handle an arbitrary number of dimensions + measures?
 * But for now, I'll only allow either:
 * 	* 2 Dimenions, 1 Measure
 *  * 1 Dimension, N Measures
 */
func (r *result) LineChart() (*[]DataSet, error) {
	numDimensions := len(r.query.dimensions)
	numMeasures := len(r.query.measures)

	// Need at least one dimension and measure
	if !(numDimensions > 0 && numMeasures > 0) {
		return nil, ErrResultMalformed
	}

	// More than 2 dimensions is straight out
	if numDimensions > 2 {
		return nil, ErrResultMalformed
	}

	// If 2 Dimensions, I can only accept 1 measure
	if numDimensions == 2 && numMeasures != 1 {
		return nil, ErrResultMalformed
	}

	var datasets []DataSet
	dimensionMap := make(map[string]int)
	curIndex := 0
	if numDimensions == 1 {
		datasets = make([]DataSet, numMeasures)
	}

	for r.rows.Next() {
		row := Row{
			Dimensions: make([]string, numDimensions),
			Measures:   make([]float64, numMeasures),
		}

		scanVals := make([]any, numDimensions+numMeasures)
		for i := range row.Dimensions {
			scanVals[i] = &row.Dimensions[i]
		}
		for i := range row.Measures {
			scanVals[numDimensions+i] = &row.Measures[i]
		}

		if err := r.rows.Scan(scanVals...); err != nil {
			return nil, err
		}

		if numDimensions > 1 {
			val, ok := dimensionMap[row.Dimensions[1]]
			if !ok {
				dimensionMap[row.Dimensions[1]] = curIndex
				val = curIndex
				curIndex++

				datasets = append(datasets, DataSet{})
			}
			datasets[val].Data = append(datasets[val].Data, dataPoint{
				X: row.Dimensions[0],
				Y: row.Measures[0],
			})
		} else {
			for i, val := range row.Measures {
				datasets[i].Data = append(datasets[i].Data, dataPoint{
					X: row.Dimensions[0],
					Y: val,
				})
			}
		}

	}
	if err := r.rows.Err(); err != nil {
		return nil, err
	}

	return &datasets, nil
}
