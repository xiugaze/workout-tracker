package main

import (
    //"database/sql"
    "encoding/json"
    "html/template"
    "log"
    "net/http"
)

func addEndpoints() {
    http.HandleFunc("/add-workout", addWorkoutHandler) 
    http.HandleFunc("/list-workouts", listWorkouts) 
    http.HandleFunc("/serve-workouts", serveWorkouts) 
    http.HandleFunc("/delete-workout", deleteWorkoutHandler)
}

type Workout struct {
    Name        string  `json:"name"`
    Duration    int     `json:"duration"`
}

type DeleteRequest struct {
    ID string `json:"id"`
}

func addWorkoutHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse JSON body
    var workout Workout
    err := json.NewDecoder(r.Body).Decode(&workout)
    if err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("INSERT INTO workouts (name, duration) VALUES ($1, $2)", workout.Name, workout.Duration)
    if err != nil {
        http.Error(w, "Error adding workout", http.StatusInternalServerError)
        return
    }

    log.Printf("Workout Added: Name=%s, Duration=%d minutes", workout.Name, workout.Duration)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message": "Workout added successfully!"}`))
}



func deleteWorkoutHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var deleteReq DeleteRequest
    err := json.NewDecoder(r.Body).Decode(&deleteReq)
    if err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("DELETE FROM workouts WHERE id = $1", deleteReq.ID)

    if err != nil {
        http.Error(w, "Error deleting workout", http.StatusInternalServerError)
        return
    }

    log.Printf("Workout Deleted: ID=%s", deleteReq.ID)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message": "Workout deleted successfully!"}`))
}


type WorkoutRecord struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Duration int    `json:"duration"`
    Time     string `json:"time"`
}

func listWorkouts(w http.ResponseWriter, _ *http.Request) {
    rows, err := db.Query("SELECT id, name, duration, time FROM workouts ORDER BY time DESC")
    if err != nil {
        http.Error(w, "Error fetching workout records", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var workoutRecords []WorkoutRecord
    for rows.Next() {
        var workoutRecord WorkoutRecord
        if err := rows.Scan(&workoutRecord.ID, &workoutRecord.Name, &workoutRecord.Duration, &workoutRecord.Time); err != nil {
            http.Error(w, "Error scanning workout records", http.StatusInternalServerError)
            return
        }
        workoutRecords = append(workoutRecords, workoutRecord)
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(workoutRecords); err != nil {
        http.Error(w, "Error serializing workout records", http.StatusInternalServerError)
        return
    }
}

/* endpoint: workouts */
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

