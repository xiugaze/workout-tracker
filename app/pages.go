package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
)

func addPages() {
    http.HandleFunc("/", serveHome)                 // Home page
    http.HandleFunc("/weeks", weeksPageHandler)     // View all weeks
    http.HandleFunc("/add-form", addFormHandler)    // Adding a workout
    http.HandleFunc("/delete-button/", deleteButtonHandler) // Deleting a workout
    http.HandleFunc("/workouts", workoutsPageHandler) // Workout list page
    http.HandleFunc("/lifts", liftsPageHandler)     // Lifts for a workout
	http.HandleFunc("/days", daysPageHandler)  
	http.HandleFunc("/meals", mealsPageHandler)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static")))) // Static files
}


func daysPageHandler(w http.ResponseWriter, r *http.Request) {
    weekID := r.URL.Query().Get("week_id")
    if weekID == "" {
        http.Error(w, "Missing week_id", http.StatusBadRequest)
        return
    }

    log.Printf("Fetching days for week_id: %s", weekID)

    rows, err := db.Query("SELECT id, week_id, day_date FROM days WHERE week_id = $1 ORDER BY day_date ASC", weekID)
    if err != nil {
        log.Printf("Error executing query: %v", err)
        http.Error(w, "Error fetching days", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var days []Day
    for rows.Next() {
        var day Day
        if err := rows.Scan(&day.ID, &day.WeekID, &day.DayDate); err != nil {
            log.Printf("Error scanning row: %v", err)
            http.Error(w, "Error scanning days", http.StatusInternalServerError)
            return
        }
        days = append(days, day)
    }

    var weekStartDate string
    err = db.QueryRow("SELECT start_date FROM weeks WHERE id = $1", weekID).Scan(&weekStartDate)
    if err != nil {
        log.Printf("Error fetching week start date: %v", err)
        http.Error(w, "Error fetching week start date", http.StatusInternalServerError)
        return
    }

    tmpl := template.Must(template.ParseFiles("templates/days.html"))
    err = tmpl.Execute(w, struct {
        WeekID        string
        WeekStartDate string
        Days          []Day
    }{
        WeekID:        weekID,
        WeekStartDate: weekStartDate,
        Days:          days,
    })
    if err != nil {
        log.Printf("Error rendering template: %v", err)
        http.Error(w, "Error rendering template", http.StatusInternalServerError)
    }
}


func weeksPageHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, start_date FROM weeks ORDER BY start_date DESC")
    if err != nil {
        http.Error(w, "Error fetching weeks", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var weeks []Week
    for rows.Next() {
        var week Week
        if err := rows.Scan(&week.ID, &week.StartDate); err != nil {
            http.Error(w, "Error scanning weeks", http.StatusInternalServerError)
            return
        }
        weeks = append(weeks, week)
    }

    tmpl := template.Must(template.ParseFiles("templates/weeks.html"))
    if err := tmpl.Execute(w, weeks); err != nil {
        http.Error(w, "Error rendering template", http.StatusInternalServerError)
    }
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func addFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	durationStr := r.FormValue("duration")
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		http.Error(w, "Error: duration is not a number", http.StatusBadRequest)
		return
	}

	workout := Workout{
		Name:     name,
		Duration: duration,
	}

	jsonData, err := json.Marshal(workout)
	if err != nil {
		http.Error(w, "Error serializing workout data", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://localhost:8080/add-workout", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Error calling add-workout endpoint", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/?success=true&workout=%s", workout.Name),
			http.StatusSeeOther,
		)
	} else {
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func deleteButtonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/delete-button/"):]

	deleteReq := struct {
		ID string `json:"id"`
	}{ID: id}

	jsonData, err := json.Marshal(deleteReq)
	if err != nil {
		http.Error(w, "Error serializing request data", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://localhost:8080/delete-workout", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		http.Redirect(w, r, "/workouts", http.StatusSeeOther)
	} else {
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func workoutsPageHandler(w http.ResponseWriter, r *http.Request) {
    dayID := r.URL.Query().Get("day_id")
    if dayID == "" {
        http.Error(w, "Missing day_id", http.StatusBadRequest)
        return
    }

    // Fetch DayDate and WeekID for the given day
    var dayDate string
    var weekID string
    err := db.QueryRow("SELECT day_date, week_id FROM days WHERE id = $1", dayID).Scan(&dayDate, &weekID)
    if err != nil {
        log.Printf("Error fetching day details: %v", err)
        http.Error(w, "Error fetching day details", http.StatusInternalServerError)
        return
    }

    // Fetch workouts for the given day
    rows, err := db.Query("SELECT id, name, duration FROM workouts WHERE day_id = $1", dayID)
    if err != nil {
        log.Printf("Error fetching workouts: %v", err)
        http.Error(w, "Error fetching workouts", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var workouts []Workout
    for rows.Next() {
        var workout Workout
        if err := rows.Scan(&workout.ID, &workout.Name, &workout.Duration); err != nil {
            log.Printf("Error scanning workouts: %v", err)
            http.Error(w, "Error processing workouts", http.StatusInternalServerError)
            return
        }
        workouts = append(workouts, workout)
    }

    // Render the workouts.html template
    tmpl := template.Must(template.ParseFiles("templates/workouts.html"))
    err = tmpl.Execute(w, struct {
        DayID    string
        DayDate  string
        WeekID   string
        Workouts []Workout
    }{
        DayID:    dayID,
        DayDate:  dayDate,
        WeekID:   weekID,
        Workouts: workouts,
    })
    if err != nil {
        log.Printf("Error rendering template: %v", err)
        http.Error(w, "Error rendering template", http.StatusInternalServerError)
    }
}


func liftsPageHandler(w http.ResponseWriter, r *http.Request) {
	workoutID := r.URL.Query().Get("workout_id")
	if workoutID == "" {
		http.Error(w, "Missing workout_id", http.StatusBadRequest)
		return
	}

	// Fetch the workout name
	var workoutName string
	db.QueryRow("SELECT name FROM workouts WHERE id = $1", workoutID).Scan(&workoutName)

	// Fetch lifts for the workout
	resp, err := http.Get("http://localhost:8080/list-lifts?workout_id=" + workoutID)
	if err != nil {
		http.Error(w, "Error fetching lifts", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var lifts []Lift
	if err := json.NewDecoder(resp.Body).Decode(&lifts); err != nil {
		http.Error(w, "Error parsing JSON data", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/lifts.html"))
	tmpl.Execute(w, struct {
		WorkoutID   string
		WorkoutName string
		Lifts       []Lift
	}{
		WorkoutID:   workoutID,
		WorkoutName: workoutName,
		Lifts:       lifts,
	})
}

func mealsPageHandler(w http.ResponseWriter, r *http.Request) {
    dayID := r.URL.Query().Get("day_id")
    if dayID == "" {
        http.Error(w, "Missing day_id", http.StatusBadRequest)
        return
    }

    var dayDate string
    var weekID string
    err := db.QueryRow("SELECT day_date, week_id FROM days WHERE id = $1", dayID).Scan(&dayDate, &weekID)
    if err != nil {
        log.Printf("Error fetching day details: %v", err)
        http.Error(w, "Error fetching day details", http.StatusInternalServerError)
        return
    }

    rows, err := db.Query("SELECT id, name, calories FROM meals WHERE day_id = $1", dayID)
    if err != nil {
        log.Printf("Error fetching meals: %v", err)
        http.Error(w, "Error fetching meals", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var meals []Meal
    for rows.Next() {
        var meal Meal
        if err := rows.Scan(&meal.ID, &meal.Name, &meal.Calories); err != nil {
            log.Printf("Error scanning meals: %v", err)
            http.Error(w, "Error processing meals", http.StatusInternalServerError)
            return
        }
        meals = append(meals, meal)
    }

    tmpl := template.Must(template.ParseFiles("templates/meals.html"))
    err = tmpl.Execute(w, struct {
        DayID    string
        DayDate  string
        WeekID   string
        Meals    []Meal
    }{
        DayID:    dayID,
        DayDate:  dayDate,
        WeekID:   weekID,
        Meals:    meals,
    })
    if err != nil {
        log.Printf("Error rendering template: %v", err)
        http.Error(w, "Error rendering template", http.StatusInternalServerError)
    }
}
