#!/bin/sh

ENDPOINTS=(
  "/add-workout"
  "/list-workouts"
  "/add-lift"
  "/list-lifts"
  "/add-workout"
  "/add-day"
  "/add-week"
  "/add-meal"
)

NUM_REQUESTS=${1:-100}
CONCURRENCY=${2:-10}
OUTPUT_DIR="results"

mkdir -p "$OUTPUT_DIR"

for ENDPOINT in "${ENDPOINTS[@]}"; do
  HOSTNAME=$(echo "$ENDPOINT" | awk -F[/:] '{print $4}')
  LOG_FILE="$OUTPUT_DIR/${ENDPOINT}_$(date +%Y%m%d%H%M%S).log"
  ab -n "$NUM_REQUESTS" -c "$CONCURRENCY" http://workout.andreano.dev/"$ENDPOINT" > "$LOG_FILE"
done
