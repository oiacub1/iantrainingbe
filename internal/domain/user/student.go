package user

type Student struct {
	User
	TrainerID string          `json:"trainerId"`
	Metadata  StudentMetadata `json:"metadata"`
}

type StudentMetadata struct {
	Goals        []string `json:"goals"`
	Injuries     []string `json:"injuries"`
	FitnessLevel string   `json:"fitnessLevel"`
	Weight       float64  `json:"weight"`
	Height       float64  `json:"height"`
	Age          int      `json:"age"`
}

const (
	FitnessLevelBeginner     = "BEGINNER"
	FitnessLevelIntermediate = "INTERMEDIATE"
	FitnessLevelAdvanced     = "ADVANCED"
)
