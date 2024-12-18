alias b := build
alias c := compile

build:
  docker compose up --build

# -C is change to directory before running
compile: 
  go build -C ./app -o app.o

# port 80 is http, so you can access just through http://localhost
up: 
  docker compose up

down: 
  docker compose down

stop:
  docker compose stop

start:
  docker compose start

clean: 
  docker compose down
  - docker rm $(docker ps -a -q)
  docker volume rm workout-tracker_db_data
