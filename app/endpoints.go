package main

import (
	"encoding/json"
	"log"
	"net/http"
	//"strconv"
)

func addEndpoints() {
	http.HandleFunc("/add-workout", addWorkoutHandler)
	http.HandleFunc("/list-workouts", listWorkoutsHandler)
	http.HandleFunc("/add-lift", addLiftHandler)
	http.HandleFunc("/list-lifts", listLiftsHandler)
    http.HandleFunc("/delete-workout", deleteWorkoutHandler)
}

type Workout struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Time     string `json:"time"`
}

type Lift struct {
    WorkoutID int     `json:"workout_id"`
    Name      string  `json:"name"`
    Weight    float64 `json:"weight"`
    Reps      int     `json:"reps"`
    LiftOrder int     `json:"lift_order"`
    RestTime  int     `json:"rest_time"`
    BPM       int     `json:"bpm"`
}

func addWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

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
	w.WriteHeader(http.StatusOK)
}

func listWorkoutsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, duration, time FROM workouts ORDER BY time DESC")
	if err != nil {
		http.Error(w, "Error fetching workouts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var workout Workout
		if err := rows.Scan(&workout.ID, &workout.Name, &workout.Duration, &workout.Time); err != nil {
			http.Error(w, "Error scanning workouts", http.StatusInternalServerError)
			return
		}
		workouts = append(workouts, workout)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workouts)
}

func addLiftHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var lift Lift
    err := json.NewDecoder(r.Body).Decode(&lift)
    if err != nil {
        log.Printf("Error decoding JSON: %v", err)
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    log.Printf("Received Lift: %+v", lift)

	_, err = db.Exec(`
        INSERT INTO lifts (workout_id, name, weight, reps, lift_order, rest_time, bpm)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		lift.WorkoutID, lift.Name, lift.Weight, lift.Reps, lift.LiftOrder, lift.RestTime, lift.BPM)
	if err != nil {
		http.Error(w, "Error adding lift", http.StatusInternalServerError)
		return
	}

	log.Printf("Lift Added: %v", lift)
	w.WriteHeader(http.StatusOK)
}

func listLiftsHandler(w http.ResponseWriter, r *http.Request) {
	workoutID := r.URL.Query().Get("workout_id")
	if workoutID == "" {
		http.Error(w, "Missing workout_id", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
        SELECT id, name, weight, reps, lift_order, rest_time, bpm
        FROM lifts WHERE workout_id = $1 ORDER BY lift_order`, workoutID)
	if err != nil {
		http.Error(w, "Error fetching lifts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lifts []Lift
	for rows.Next() {
		var lift Lift
		if err := rows.Scan(&lift.WorkoutID, &lift.Name, &lift.Weight, &lift.Reps, &lift.LiftOrder, &lift.RestTime, &lift.BPM); err != nil {
			http.Error(w, "Error scanning lifts", http.StatusInternalServerError)
			return
		}
		lifts = append(lifts, lift)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lifts)
}

func deleteWorkoutHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        ID string `json:"id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    _, err := db.Exec("DELETE FROM workouts WHERE id = $1", req.ID)
    if err != nil {
        http.Error(w, "Error deleting workout", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}