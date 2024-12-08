package main 

import (
    "os"
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"

    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
    var err error
    db, err = sql.Open("sqlite3", "./workouts.db")
    if err != nil {
        log.Fatal(err)
    }

    createTableQuery := `
    CREATE TABLE IF NOT EXISTS workouts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        duration INTEGER NOT NULL,
        time DATETIME DEFAULT CURRENT_TIMESTAMP
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

    _, err := db.Exec("INSERT INTO workouts (name, duration) VALUES (?, ?)", name, duration)
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

    _, err := db.Exec("DELETE FROM workouts WHERE id = ?", id)
    if err != nil {
        http.Error(w, "Error deleting workout", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/workouts", http.StatusSeeOther)
}
