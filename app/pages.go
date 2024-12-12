package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
)

func addPages() {
	http.HandleFunc("/", serveHome)                  // Home page
	http.HandleFunc("/add-form", addFormHandler)    // Adding a workout
	http.HandleFunc("/delete-button/", deleteButtonHandler) // Deleting a workout
	http.HandleFunc("/workouts", workoutsPageHandler) // Workout list page
	http.HandleFunc("/lifts", liftsPageHandler)       // Lifts for a workout
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static")))) // Static files
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
	resp, err := http.Get("http://localhost:8080/list-workouts")
	if err != nil {
		http.Error(w, "Error fetching workout records", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var workoutRecords []Workout
	if err := json.NewDecoder(resp.Body).Decode(&workoutRecords); err != nil {
		http.Error(w, "Error parsing JSON data", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/workouts.html"))
	if err := tmpl.Execute(w, workoutRecords); err != nil {
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
