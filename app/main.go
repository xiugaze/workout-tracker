package main 

import (
    "os"
    "database/sql"
    "fmt"
    //"html/template"
    "log"
    "net/http"
    // _ "github.com/mattn/go-sqlite3"
    _ "github.com/lib/pq"
)

var db *sql.DB

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Please provide an IP address")
        return
    }
    ip := os.Args[1]

    initDB()
    defer db.Close()

    addPages()
    addEndpoints()

    fmt.Println("Server is running on port 8080")
    http.ListenAndServe(ip + ":8080", nil)
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}

func initDB() {
    var err error

    connStr := fmt.Sprintf(
        // https://stackoverflow.com/questions/26343178/dockerizing-postgresql-psql-connection-refused
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        getEnv("DB_HOST", "host.docker.internal"),
        getEnv("DB_PORT", "5432"),
        getEnv("DB_USER", "postgres"),
        getEnv("DB_PASSWORD", "password"),
        getEnv("DB_NAME", "testdb"),
    )

    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    createWorkoutsTable()
    createAnalyticsTable()
}


func createAnalyticsTable() {
    var err error
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS analytics (
        id SERIAL PRIMARY KEY,
        time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        endpoint TEXT NOT NULL,
        data TEXT NOT NULL
    );`
    _, err = db.Exec(createTableQuery)
    if err != nil {
        log.Fatal(err)
    }
}

func createWorkoutsTable() {
    var err error
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS workouts (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        duration INTEGER NOT NULL,
        time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
    _, err = db.Exec(createTableQuery)
    if err != nil {
        log.Fatal(err)
    }
}

