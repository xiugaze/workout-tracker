package main

import (
    "bytes"
    "encoding/json"
    "html/template"
    "io"
    "net/http"
    "strconv"
)

func addPages() {
    http.HandleFunc("/", serveHome) // home page 
    http.HandleFunc("/add-form", addFormHandler) // adding workout form
    http.HandleFunc("/delete-button/", deleteButtonHandler) // adding workout form
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static")))) // CSS
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
        return;
    }
    name := r.FormValue("name")
    durationStr := r.FormValue("duration")
    duration, err:= strconv.Atoi(durationStr)
    if err != nil {
        http.Error(w, "Error: duration is not a number", http.StatusBadRequest)
        return;
    }

    workout := Workout {
        Name: name,
        Duration: duration,
    }

    jsonData, err := json.Marshal(workout)
    if err != nil {
        http.Error(w, "Error serializing workout data", http.StatusInternalServerError)
        return;
    }

    resp, err := http.Post("http://localhost:8080/add-workout", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        http.Error(w, "Error calling add-workout endpoint", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Relay the response back to the client
    w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}

func deleteButtonHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    id := r.URL.Path[len("/delete-button/"):]

    deleteReq := DeleteRequest{ID: id}
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

    w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}
