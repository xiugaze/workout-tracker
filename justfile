alias b := build
alias c := compile

build:
  docker build -t workout-tracker .

compile: 
  go build .

# port 80 is http, so you can access just through http://localhost
run: 
  docker run -p 80:8080 workout-tracker 
