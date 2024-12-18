package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	// "io/ioutil"
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
    http.HandleFunc("/add-lift-button", addLiftButtonHandler)
    http.HandleFunc("/analytics", analyticsHandler)
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


// type Lift struct {
//     WorkoutID int     `json:"workout_id"`
//     Name      string  `json:"name"`
//     Weight    float64 `json:"weight"`
//     Reps      int     `json:"reps"`
//     LiftOrder int     `json:"lift_order"`
//     RestTime  int     `json:"rest_time"`
//     BPM       int     `json:"bpm"`
// }
func addLiftButtonHandler(w http.ResponseWriter, r *http.Request) {
	log.SetOutput(os.Stdout)
	log.Println("got here")
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	// 	return
	// }
	//
	// // Parse form data
	// err := r.ParseForm()
	// if err != nil {
	// 	http.Error(w, "Error parsing form data", http.StatusBadRequest)
	// 	return
	// }
	//
	// // Convert form fields to the Lift struct
	// lift := Lift{}
	//
	// lift.WorkoutID, err = strconv.Atoi(r.FormValue("workout_id"))
	// if err != nil || lift.WorkoutID <= 0 {
	// 	http.Error(w, "Invalid workout ID", http.StatusBadRequest)
	// 	return
	// }
	//
	// lift.Name = r.FormValue("name")
	// if lift.Name == "" {
	// 	http.Error(w, "Missing lift name", http.StatusBadRequest)
	// 	return
	// }
	//
	// lift.Weight, err = strconv.ParseFloat(r.FormValue("weight"), 64)
	// if err != nil || lift.Weight <= 0 {
	// 	http.Error(w, "Invalid weight", http.StatusBadRequest)
	// 	return
	// }
	//
	// lift.Reps, err = strconv.Atoi(r.FormValue("reps"))
	// if err != nil || lift.Reps <= 0 {
	// 	http.Error(w, "Invalid reps", http.StatusBadRequest)
	// 	return
	// }
	//
	// lift.LiftOrder, err = strconv.Atoi(r.FormValue("lift_order"))
	// if err != nil || lift.LiftOrder <= 0 {
	// 	http.Error(w, "Invalid lift order", http.StatusBadRequest)
	// 	return
	// }
	//
	// lift.RestTime, err = strconv.Atoi(r.FormValue("rest_time"))
	// if err != nil {
	// 	lift.RestTime = 0 // Default to 0 if not provided
	// }
	//
	// lift.BPM, err = strconv.Atoi(r.FormValue("bpm"))
	// if err != nil {
	// 	lift.BPM = 0 // Default to 0 if not provided
	// }
	//
	//
	//
	//    _, err = db.Exec(
	// "INSERT INTO lifts (workout_id, name, weight, reps, lift_order, rest_time, bpm) VALUES ($1, $2, $3, $4, $5, $6, $7)",
	// lift.WorkoutID, lift.Name, lift.Weight, lift.Reps, lift.LiftOrder, lift.RestTime, lift.BPM,
	//    )
	//    if err != nil {
	// log.Printf("Error inserting lift: %v", err)
	// http.Error(w, "Error adding lift", http.StatusInternalServerError)
	// return
	//    }


	// jsonData, err := json.Marshal(lift)
	// if err != nil {
	// 	http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	// 	return
	// }
	//
	// resp, err := http.Post("http://localhost:8080/add-lift", "application/json", bytes.NewBuffer(jsonData))
	//
	// if err != nil {
	// 	http.Error(w, "Error calling addLiftHandler", http.StatusInternalServerError)
	// 	return
	// }
	// defer resp.Body.Close()
	//
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	http.Error(w, "Error reading response", http.StatusInternalServerError)
	// 	return
	// }
	//
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(resp.StatusCode)
	// w.Write(body)
	//
	// if resp.StatusCode == http.StatusOK {
	// 	fmt.Println("got here")
	// 	http.Redirect(
	// 		w,
	// 		r,
	// 		fmt.Sprintf("/lifts?workout_id=%v", lift.WorkoutID),
	// 		http.StatusSeeOther,
	// 	)
	// } else {
	// 	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	// 	w.WriteHeader(resp.StatusCode)
	// 	io.Copy(w, resp.Body)
	// }

}

func analyticsHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT endpoint, visit_count FROM endpoint_visits ORDER BY endpoint ASC")
    if err != nil {
        log.Printf("Error fetching endpoint visits: %v", err)
        http.Error(w, "Error fetching data", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var visits []struct {
        Endpoint   string
        VisitCount int
    }
    for rows.Next() {
        var endpoint string
        var visitCount int
        if err := rows.Scan(&endpoint, &visitCount); err != nil {
            log.Printf("Error scanning row: %v", err)
            http.Error(w, "Error processing data", http.StatusInternalServerError)
            return
        }
        visits = append(visits, struct {
            Endpoint   string
            VisitCount int
        }{
            Endpoint:   endpoint,
            VisitCount: visitCount,
        })
    }

    // Render the HTML page
    tmpl := template.Must(template.New("endpoint_visits").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Endpoint Visit Counts</title>
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <div class="container">
        <h1>Endpoint Visit Counts</h1>
        <table>
            <thead>
                <tr>
                    <th>Endpoint</th>
                    <th>Visit Count</th>
                </tr>
            </thead>
            <tbody>
                {{range .}}
                <tr>
                    <td>{{.Endpoint}}</td>
                    <td>{{.VisitCount}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        <a href="/"><button>Back to Home</button></a>
    </div>
</body>
</html>
    `))

    // Execute the template with the visits data
    if err := tmpl.Execute(w, visits); err != nil {
        log.Printf("Error rendering template: %v", err)
        http.Error(w, "Error rendering page", http.StatusInternalServerError)
    }
}





