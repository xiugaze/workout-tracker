alias b := build
alias c := compile

build:
  docker compose up --build

compile: 
  go build ./app

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
  docker rm $(docker ps -a -q)
