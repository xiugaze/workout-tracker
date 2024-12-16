package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var tmpl = template.Must(template.New("addCol").Parse(`
    <div style="display: flex; flex-direction: row; gap: 10px;">
      <div>
        <label for="weight{{.Counter}}">Weight</label><br>
        <input type="number" id="weight{{.Counter}}" name="weight{{.Counter}}" required>
      </div>
      <div>
        <label for="reps{{.Counter}}">Reps</label><br>
        <input type="number" id="reps{{.Counter}}" name="reps{{.Counter}}" required>
      </div>
    </div>
    <br>
`))

func main() {
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/", fs)
    http.HandleFunc("/addCol", addCol)
    http.HandleFunc("/submit", submit)

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Running on 8080")
}

func addCol(w http.ResponseWriter, r *http.Request) { 
    counter := r.URL.Query().Get("counter")
    log.Printf("counter: %v", counter)
    if counter == "" {
        http.Error(w, "Counter is required", http.StatusBadRequest)
        log.Printf("Counter is required")
        return
    }

    err := tmpl.Execute(w, map[string]string{"Counter": counter})
    if err != nil {
        http.Error(w, "Error rendering template", http.StatusInternalServerError)
        log.Printf("Error rendering template")
    }
}


func submit(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Error parsing form data", http.StatusInternalServerError)
        return
    }

    for key, values := range r.Form {
        for _, value := range values {
            fmt.Printf("%s: %s\n", key, value)
        }
    }

    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, `
    <p>Thank you for your submission!</p>
`)
}
