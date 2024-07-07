package main

import (
	"database/sql"
	"fmt"
	"ga4ct/event"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func connctDB() {
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
}

func makeDummyData(numUsers int, avgSessionsPerUser float64, avgEventsPerSession float64,
	startTime time.Time, endTime time.Time) *[]event.Event {

	accountId := "GA4CT-1"
	hostname := "https://www.example.com"
	// name := []string{
	// 	"page_view",
	// 	"cta_click",
	// 	"form_submission",
	// 	"purchase",
	// }

	pages := []struct {
		Path  string
		Title string
	}{
		{"/", "Storefront | We Sell Products | Example"},
		{"/collection", "Product Collection | Example"},
		{"/product-a", "Product A | Example"},
		{"/product-b", "Product B | Example"},
		{"/product-c", "Product C | Example"},
		{"/cart", "Your Cart | Example"},
		{"/contact", "Contact Us | Example"},
		{"/checkout", "Checkout | Example"},
	}

	// purchase_page := "/order-confirmation"
	// lead_form_page := "/contact-thank-you"

	agents := []struct {
		UserAgent        string
		ScreenResolution string
	}{
		// Chrome, Desktop, Windows
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
			"1920x1080",
		},
		// Chrome, Mobile, Android / Pixel 7
		{
			"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Mobile Safari/537.36",
			"412x915",
		},
		// Safari, Tablet, iOS / iPad Pro
		{
			"Mozilla/5.0 (iPad; CPU OS 17_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
			"1024x1366",
		},
		// Safari, Mobile, iOS / iPhone
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
			"390x844",
		},
	}

	type Channels int

	const (
		Direct = iota
		OrganicSearch
		OrganicSocial
		Email
		Referral
		PaidSearch
		PaidSocial
	)

	channels := []Channels{
		Direct,
		OrganicSearch,
		OrganicSocial,
		Email,
		Referral,
		PaidSearch,
		PaidSocial,
	}

	channelsMap := map[Channels]struct {
		referrers []string
		query     []string
	}{
		Direct: {
			referrers: []string{""},
			query:     []string{""},
		},
		OrganicSearch: {
			referrers: []string{
				"https://www.google.com/",
				"https://www.bing.com/",
				"https://www.duckduckgo.com/",
			},
			query: []string{""},
		},
		OrganicSocial: {
			referrers: []string{
				"https://www.facebook.com/",
				"https://www.instagram.com/",
				"https://www.pinterest.com/",
				"https://www.reddit.com/",
				"https://www.youtube.com/",
			},
			query: []string{""},
		},
		Email: {
			referrers: []string{
				"https://mail.google.com/",
				"https://outlook.live.com/",
				"https://mail.yahoo.com/",
			},
			query: []string{""},
		},
		Referral: {
			referrers: []string{
				"https://www.referrer1.com/",
				"https://www.referrer2.com/",
				"https://www.referrer3.com/",
			},
			query: []string{""},
		},
		PaidSearch: {
			referrers: []string{
				"https://www.google.com/",
			},
			query: []string{
				"?utm_source=google&utm_medium=cpc&utm_campaign=PMax%20-%20All%20Products",
				"?utm_source=google&utm_medium=cpc&utm_campaign=Search%20-%20Brand",
				"?utm_source=google&utm_medium=cpc&utm_campaign=Search%20-%20Non-Brand",
			},
		},
		PaidSocial: {
			referrers: []string{
				"https://www.facebook.com/",
			},
			query: []string{
				"?utm_source=facebook&utm_medium=paid+social&utm_campaign=Prospecting%20-%20Video",
				"?utm_source=facebook&utm_medium=paid+social&utm_campaign=Remarketing%20-%20All%20Website%20Visitors",
			},
		},
	}

	type User struct {
		ID               string
		FirstSessionTime time.Time
	}

	type Session struct {
		ID               string
		SessionStartTime time.Time
	}

	users := make([]User, numUsers)

	totalSessionsCount := int(float64(numUsers) * avgSessionsPerUser)
	sessions := make([]Session, totalSessionsCount)

	totalEventCount := int(float64(numUsers) * avgSessionsPerUser * avgEventsPerSession)
	events := make([]event.Event, totalEventCount)

	// Add the first landing page views for every user
	for i := 0; i < numUsers; i++ {
		timestamp := randomTimeBetween(startTime, endTime)
		users[i] = User{ID: "GA4CT.CID." + generateRandomID(timestamp), FirstSessionTime: timestamp}
		sessions[i] = Session{ID: "GA4CT.SID." + generateRandomID(timestamp), SessionStartTime: timestamp}

		page := pages[rand.Intn(len(pages))]
		agent := agents[rand.Intn(len(agents))]
		channel := channels[rand.Intn(len(channels))]
		referrer := channelsMap[channel].referrers[rand.Intn(len(channelsMap[channel].referrers))]
		query := channelsMap[channel].query[rand.Intn(len(channelsMap[channel].query))]

		events[i] = event.Event{
			AccountID:        accountId,
			ClientID:         users[i].ID,
			SessionID:        sessions[i].ID,
			Name:             "page_view",
			Value:            0,
			Timestamp:        timestamp,
			PageLocation:     hostname + page.Path + query,
			PageTitle:        page.Title,
			PageReferrer:     referrer,
			UserAgent:        agent.UserAgent,
			ScreenResolution: agent.ScreenResolution,
		}
	}

	// Add the first landing page views for repeat sessions
	// Repeat sessions need to come after first user session

	// Fill the remaining event space with random events
	// Events in the same session need to be within 30 minutes of the first event in the session

	return &events
}

func randomTimeBetween(startTime time.Time, endTime time.Time) time.Time {
	duration := endTime.Sub(startTime)
	randomDuration := time.Duration(rand.Int63n(int64(duration)))
	return startTime.Add(randomDuration)
}

func generateRandomID(timestamp time.Time) string {
	randomPart := strconv.Itoa(int(rand.Int31() & 0x7FFFFFFF))
	timestampPart := strconv.FormatInt(timestamp.Unix(), 10)
	return randomPart + "." + timestampPart
}

func main() {
	rand.Seed(time.Now().UnixNano())

	events := makeDummyData(1, 2.5, 5.15, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 12, 31, 23, 59, 99, 0, time.UTC))

	fmt.Printf("%v\n", *events)
}
