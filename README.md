# Workout Tracker
#### Christian Basso and Caleb Andreano

This application is a custom workout tracker written in Go that allows users to categorize their workouts and meals by day and week.

## Getting Started
### Prereuisites
- Go version 1.16 or higher.
- PostgreSQL database.

With the just implimentaion, all you need to do to lauch the app is run the command

```
just build
```

Then the container should build itself automatically.

## API Documentation

### Endpoints

#### Week Management
- **Add Week**
  - **POST** `/add-week`
  - **Payload:** `{ "start_date": "YYYY-MM-DD" }`
- **View Weeks**
  - **GET** `/weeks`

#### Day Management
- **Add Day**
  - **POST** `/add-day`
  - **Payload:** `{ "week_id": 1, "day_date": "YYYY-MM-DD" }`
- **View Days**
  - **GET** `/days?week_id=<WEEK_ID>`

#### Workout Management
- **Add Workout**
  - **POST** `/add-workout`
  - **Payload:** `{ "day_id": 1, "name": "Workout Name", "duration": 60 }`
- **List Workouts**
  - **GET** `/list-workouts?day_id=<DAY_ID>`

#### Lift Management
- **Add Lift**
  - **POST** `/add-lift`
  - **Payload:** `{ "workout_id": 1, "name": "Lift Name", "weight": 100.0, "reps": 10, "rest_time": 60 }`
- **List Lifts**
  - **GET** `/list-lifts?workout_id=<WORKOUT_ID>`

#### Meal Management
- **Add Meal**
  - **POST** `/add-meal`
  - **Payload:** `{ "day_id": 1, "name": "Meal Name", "calories": 300 }`

#### Analytics
- **View Endpoint Visits**
  - **GET** `/analytics`

---

## Database Schema

### Weeks
```sql
CREATE TABLE weeks (
    id SERIAL PRIMARY KEY,
    start_date DATE NOT NULL
);
```

### Days
```sql
CREATE TABLE days (
    id SERIAL PRIMARY KEY,
    week_id INTEGER REFERENCES weeks(id),
    day_date DATE NOT NULL
);
```

### Workout
```sql
CREATE TABLE workouts (
    id SERIAL PRIMARY KEY,
    day_id INTEGER REFERENCES days(id),
    name TEXT NOT NULL,
    duration INTEGER NOT NULL
);
```

### Lifts
```sql
CREATE TABLE lifts (
    id SERIAL PRIMARY KEY,
    workout_id INTEGER REFERENCES workouts(id),
    name TEXT NOT NULL,
    weight DOUBLE PRECISION NOT NULL,
    reps INTEGER NOT NULL,
    rest_time INTEGER NOT NULL
);
```

### Meals
```sql
CREATE TABLE meals (
    id SERIAL PRIMARY KEY,
    day_id INTEGER REFERENCES days(id),
    name TEXT NOT NULL,
    calories INTEGER NOT NULL
);
```


