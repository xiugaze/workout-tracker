package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	//"strconv"
	"time"
	//"strconv"
)

func addEndpoints() {
	http.HandleFunc("/add-workout", addWorkoutHandler)
	http.HandleFunc("/list-workouts", listWorkoutsHandler)
	http.HandleFunc("/add-lift", addLiftHandler)
	http.HandleFunc("/list-lifts", listLiftsHandler)
    http.HandleFunc("/delete-workout", deleteWorkoutHandler)
	http.HandleFunc("/add-week", addWeekHandler)
    http.HandleFunc("/add-day", addDayHandler)
    http.HandleFunc("/add-meal", addMealHandler)
}

type Workout struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Time     time.Time `json:"time"`
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
type Week struct {
    ID        int    `json:"id"`
    StartDate string `json:"start_date"`
}

type Day struct {
    ID       int    `json:"id"`
    WeekID   int    `json:"week_id"`
    DayDate  string `json:"day_date"`
}

type Meal struct {
    ID       int    `json:"id"`
    DayID    int    `json:"day_id"`
    Name     string `json:"name"`
    Calories int    `json:"calories"`
}

func addWeekHandler(w http.ResponseWriter, r *http.Request) {
    incrementVisit("add-week")
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var week Week
    if err := json.NewDecoder(r.Body).Decode(&week); err != nil {
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    _, err := db.Exec("INSERT INTO weeks (start_date) VALUES ($1)", week.StartDate)
    if err != nil {
        http.Error(w, "Error adding week", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func addDayHandler(w http.ResponseWriter, r *http.Request) {
    incrementVisit("add-day")
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    weekID := r.FormValue("week_id")
    dayDate := r.FormValue("day_date")

    if weekID == "" || dayDate == "" {
        http.Error(w, "Missing week_id or day_date", http.StatusBadRequest)
        return
    }

    // Insert the new day and return the inserted ID
    var dayID int
    err := db.QueryRow("INSERT INTO days (week_id, day_date) VALUES ($1, $2) RETURNING id", weekID, dayDate).Scan(&dayID)
    if err != nil {
        log.Printf("Error inserting day: %v", err)
        http.Error(w, "Error inserting day", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "message": "Day added successfully",
        "dayID": dayID,
    })
}

func addMealHandler(w http.ResponseWriter, r *http.Request) {
    incrementVisit("add-meal")
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var meal Meal
    if err := json.NewDecoder(r.Body).Decode(&meal); err != nil {
        log.Printf("Invalid JSON format: %v", err)
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    _, err := db.Exec("INSERT INTO meals (day_id, name, calories) VALUES ($1, $2, $3)", meal.DayID, meal.Name, meal.Calories)
    if err != nil {
        log.Printf("Error inserting meal: %v", err)
        http.Error(w, "Error adding meal", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "message": "Meal added successfully",
    })
}

func addWorkoutHandler(w http.ResponseWriter, r *http.Request) {
    incrementVisit("add-workout")
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse JSON data
    var workout struct {
        DayID    string `json:"day_id"`
        Name     string `json:"name"`
        Duration int    `json:"duration"`
    }

    err := json.NewDecoder(r.Body).Decode(&workout)
    if err != nil {
        log.Printf("Invalid JSON data: %v", err)
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    if workout.DayID == "" || workout.Name == "" || workout.Duration <= 0 {
        http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
        return
    }

    // Insert workout into the database
    var workoutID int
    err = db.QueryRow("INSERT INTO workouts (day_id, name, duration) VALUES ($1, $2, $3) RETURNING id",
        workout.DayID, workout.Name, workout.Duration).Scan(&workoutID)
    if err != nil {
        log.Printf("Error inserting workout: %v", err)
        http.Error(w, "Error adding workout", http.StatusInternalServerError)
        return
    }

    // Respond with the new workout ID
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "id":     workoutID,
    })
}


func listWorkoutsHandler(w http.ResponseWriter, r *http.Request) {
	incrementVisit("list-workouts")
	dayIDStr := r.URL.Query().Get("day_id")
	if dayIDStr == "" {
		http.Error(w, "day_id parameter is required", http.StatusBadRequest)
		return
	}

	

	dayID, err := strconv.Atoi(dayIDStr)
	fmt.Printf("got %v", dayID)

	if err != nil {
		http.Error(w, "Invalid day_id parameter", http.StatusBadRequest)
		return
	}
	rows, err := db.Query("SELECT id, name, duration FROM workouts WHERE day_id = $1", dayID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		http.Error(w, "Error fetching workouts", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var workout Workout
		if err := rows.Scan(&workout.ID, &workout.Name, &workout.Duration); err != nil {
			http.Error(w, "Error scanning workouts", http.StatusInternalServerError)
			return
		}
		workouts = append(workouts, workout)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workouts)
}

func addLiftHandler(w http.ResponseWriter, r *http.Request) {
    incrementVisit("add-lift")
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var lift struct {
        WorkoutID int     `json:"workout_id"`
        Name      string  `json:"name"`
        Weight    float64 `json:"weight"`
        Reps      int     `json:"reps"`
        LiftOrder int     `json:"lift_order"`
        RestTime  int     `json:"rest_time"`
        BPM       int     `json:"bpm"`
    }

    // Decode JSON from the request body
    if err := json.NewDecoder(r.Body).Decode(&lift); err != nil {
        log.Printf("Invalid JSON format: %v", err)
        http.Error(w, "Invalid JSON data", http.StatusBadRequest)
        return
    }

    // Validate the incoming data
    if lift.WorkoutID <= 0 || lift.Name == "" || lift.Weight <= 0 || lift.Reps <= 0 || lift.LiftOrder <= 0 {
        http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
        return
    }

    // Insert the lift into the database
    _, err := db.Exec(
        "INSERT INTO lifts (workout_id, name, weight, reps, lift_order, rest_time, bpm) VALUES ($1, $2, $3, $4, $5, $6, $7)",
        lift.WorkoutID, lift.Name, lift.Weight, lift.Reps, lift.LiftOrder, lift.RestTime, lift.BPM,
    )
    if err != nil {
        log.Printf("Error inserting lift: %v", err)
        http.Error(w, "Error adding lift", http.StatusInternalServerError)
        return
    }

    // Return a JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "success",
        "message": "Lift added successfully",
    })
}



func listLiftsHandler(w http.ResponseWriter, r *http.Request) {
	incrementVisit("list-lifts")
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

func incrementVisit(endpoint string) {
    query := `
    UPDATE endpoint_visits
    SET visit_count = visit_count + 1
    WHERE endpoint = $1;
    `
    _, err := db.Exec(query, endpoint)
    if err != nil {
        log.Printf("Error incrementing visit count for %s: %v", endpoint, err)
    } else {
        log.Printf("Visit count incremented for %s", endpoint)
    }
}
