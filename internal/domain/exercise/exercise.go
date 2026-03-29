package exercise

import (
	"time"
)

type Exercise struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	NameKey       string        `json:"nameKey"`
	DescriptionKey string       `json:"descriptionKey"`
	YoutubeURL    string        `json:"youtubeUrl"`
	ThumbnailURL  string        `json:"thumbnailUrl"`
	MuscleGroups  []MuscleGroup `json:"muscleGroups"`
	Difficulty    Difficulty    `json:"difficulty"`
	Equipment     []string      `json:"equipment"`
	CreatedBy     string        `json:"createdBy"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type Difficulty string

const (
	DifficultyBeginner     Difficulty = "BEGINNER"
	DifficultyIntermediate Difficulty = "INTERMEDIATE"
	DifficultyAdvanced     Difficulty = "ADVANCED"
)

func (d Difficulty) IsValid() bool {
	switch d {
	case DifficultyBeginner, DifficultyIntermediate, DifficultyAdvanced:
		return true
	default:
		return false
	}
}
