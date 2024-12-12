package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Check if an IP address is provided
	if len(os.Args) < 2 {
		fmt.Println("Please provide an IP address")
		return
	}
	ip := os.Args[1]

	// Initialize database connection
	initDB()
	defer db.Close()

	// Add routes and start the server
	addPages()
	addEndpoints()

	log.Println("Server is running on port 8080")
	http.ListenAndServe(ip+":8080", nil)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func initDB() {
	var err error

	// Build the connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "host.docker.internal"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "testdb"),
	)

	// Connect to the database
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Create necessary tables
	createWorkoutsTable()
	createLiftsTable()
}

func createWorkoutsTable() {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS workouts (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        duration INTEGER NOT NULL,
        time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}

func createLiftsTable() {
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS lifts (
        id SERIAL PRIMARY KEY,
        workout_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        weight FLOAT NOT NULL,
        reps INTEGER NOT NULL,
        lift_order INTEGER NOT NULL,
        rest_time INTEGER NOT NULL,
        bpm INTEGER,
        FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE
    );`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
}
