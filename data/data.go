package data

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var ErrNoResult = errors.New("no results")

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

type ModeledDimension struct {
	Label string
	SQL   string
}

type ModeledMeasure struct {
	Label string
	SQL   string
}

var (
	EventName = ModeledDimension{
		Label: "Event Name",
		SQL:   `name`,
	}
	Date = ModeledDimension{
		Label: "Date",
		SQL:   `TO_CHAR(timestamp, 'YYYY-MM-DD')`,
	}
)

var (
	EventCount = ModeledMeasure{
		Label: "Event Count",
		SQL:   `COUNT(*)`,
	}
	EventValue = ModeledMeasure{
		Label: "Event Value",
		SQL:   `ROUND(SUM(value),2)`,
	}
)

type Dimension string

type Measure float64

type Row struct {
	Dimensions []Dimension
	Measures   []Measure
}

type DataPoint struct {
	Label string
	Value int
}

func RetrieveSQL2(db *sql.DB, sqlStmt string) (*[]DataPoint, error) {
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []DataPoint
	for rows.Next() {
		var dataPoint DataPoint
		if err := rows.Scan(&dataPoint.Label, &dataPoint.Value); err != nil {
			return nil, err
		}
		data = append(data, dataPoint)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &data, nil
}

func buildSQL(dimensions []ModeledDimension, measures []ModeledMeasure, limit int) string {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("SELECT ")

	if len(dimensions) > 0 && len(measures) == 0 {
		sqlBuilder.WriteString("DISTINCT ")
	}

	for i, dimension := range dimensions {
		sqlBuilder.WriteString(dimension.SQL)
		if !(i == len(dimensions)-1 && len(measures) == 0) {
			sqlBuilder.WriteString(", ")
		}
	}

	for i, measure := range measures {
		sqlBuilder.WriteString(measure.SQL)
		if i != len(measures)-1 {
			sqlBuilder.WriteString(", ")
		}
	}

	sqlBuilder.WriteString(" FROM events ")

	if len(measures) == 0 {
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

func Retrieve(db *sql.DB, modeledDimensions []ModeledDimension, modeledMeasures []ModeledMeasure) (*[]Row, error) {
	if len(modeledDimensions) == 0 && len(modeledMeasures) == 0 {
		return nil, ErrNoResult
	}

	rows, err := db.Query(buildSQL(modeledDimensions, modeledMeasures, 10))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []Row
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
		data = append(data, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &data, nil
}
