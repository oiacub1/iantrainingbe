package main

type CreateRoutineRequest struct {
	Name        string `json:"name"`
	WeekCount   int    `json:"weekCount"`
	Description string `json:"description"`
}

type UpdateRoutineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type CreateWorkoutDayRequest struct {
	WeekNumber int                    `json:"weekNumber"`
	DayNumber  int                    `json:"dayNumber"`
	DayName    string                 `json:"dayName"`
	IsRestDay  bool                   `json:"isRestDay"`
	Exercises  []CreateExerciseSetReq `json:"exercises"`
}

type CreateExerciseSetReq struct {
	ExerciseID  string `json:"exerciseId"`
	Order       int    `json:"order"`
	Sets        int    `json:"sets"`
	Reps        string `json:"reps"`
	RestSeconds int    `json:"restSeconds"`
	Notes       string `json:"notes"`
	Tempo       string `json:"tempo"`
	RPE         int    `json:"rpe"`
}
