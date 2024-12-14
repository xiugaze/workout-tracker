#!/bin/sh

ENDPOINTS=(
  "delete-workout"
  "add-workout"
  "list-workouts"
  "add-lift"
  "list-lifts"
  "add-workout"
  "add-day"
  "add-week"
  "add-meal"
)

NUM_REQUESTS=${1:-100}
CONCURRENCY=${2:-10}
OUTPUT_DIR="results"
CSV_FILE="$OUTPUT_DIR/summary.csv"

mkdir -p "$OUTPUT_DIR"
echo "Endpoint,Min,Mean,SD,Median,Max" > "$CSV_FILE"

for ENDPOINT in "${ENDPOINTS[@]}"; do
  HOSTNAME=$(echo "$ENDPOINT" | awk -F[/:] '{print $4}')
  LOG_FILE="$OUTPUT_DIR/${ENDPOINT}_$(date +%Y%m%d%H%M%S).log"
  ab -n "$NUM_REQUESTS" -c "$CONCURRENCY" "http://workout.andreano.dev/$ENDPOINT" > "$LOG_FILE"

  TOTAL_ROW=$(grep -E "^Total:" "$LOG_FILE" | awk '{print $2, $3, $4, $5, $6}')
  if [[ -n "$TOTAL_ROW" ]]; then
    MIN=$(echo "$TOTAL_ROW" | awk '{print $1}')
    MEAN=$(echo "$TOTAL_ROW" | awk '{print $2}')
    SD=$(echo "$TOTAL_ROW" | awk '{print $3}')
    MEDIAN=$(echo "$TOTAL_ROW" | awk '{print $4}')
    MAX=$(echo "$TOTAL_ROW" | awk '{print $5}')

    echo "$ENDPOINT,$MIN,$MEAN,$SD,$MEDIAN,$MAX" >> "$CSV_FILE"
  fi
done
