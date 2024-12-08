package main 

import (
    "os"
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"
    // _ "github.com/mattn/go-sqlite3"
    _ "github.com/lib/pq"
)

var db *sql.DB

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

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Please provide an IP address")
        return
    }
    ip := os.Args[1]

    initDB()
    defer db.Close()
    http.HandleFunc("/", serveHome)
    http.HandleFunc("/workout", addWorkout)
    http.HandleFunc("/workouts", serveWorkouts)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.HandleFunc("/delete/", deleteWorkout)

    fmt.Println("Server is running on port 8080")
    http.ListenAndServe(ip + ":8080", nil)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func serveHome(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/index.html"))
    tmpl.Execute(w, nil)
}

func addWorkout(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    name := r.FormValue("name")
    duration := r.FormValue("duration")

    _, err := db.Exec("INSERT INTO workouts (name, duration) VALUES ($1, $2)", name, duration)

    if err != nil {
        http.Error(w, "Error adding workout", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func serveWorkouts(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, duration, time FROM workouts ORDER BY time DESC")
    if err != nil {
        http.Error(w, "Error fetching workouts", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var workouts []struct {
        ID       int
        Name     string
        Duration int
        Time     string
    }

    for rows.Next() {
        var workout struct {
            ID       int
            Name     string
            Duration int
            Time     string
        }
        err := rows.Scan(&workout.ID, &workout.Name, &workout.Duration, &workout.Time)
        if err != nil {
            http.Error(w, "Error scanning workouts", http.StatusInternalServerError)
            return
        }
        workouts = append(workouts, workout)
    }

    tmpl := template.Must(template.ParseFiles("templates/workouts.html"))
    tmpl.Execute(w, workouts)
}

func deleteWorkout(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/delete/"):]

    _, err := db.Exec("DELETE FROM workouts WHERE id = $1", id)
    if err != nil {
        http.Error(w, "Error deleting workout", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/workouts", http.StatusSeeOther)
}
