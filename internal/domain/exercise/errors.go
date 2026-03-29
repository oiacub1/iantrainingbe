package exercise

import "errors"

var (
	ErrExerciseNotFound         = errors.New("exercise not found")
	ErrInvalidExerciseID        = errors.New("invalid exercise ID")
	ErrInvalidYoutubeURL        = errors.New("invalid YouTube URL")
	ErrInvalidImpactPercentage  = errors.New("impact percentage must be between 0 and 100")
	ErrInvalidTotalImpact       = errors.New("total muscle group impact must equal 100%")
	ErrNoMuscleGroups           = errors.New("at least one muscle group is required")
	ErrMuscleGroupRequired      = errors.New("muscle group is required")
	ErrMuscleGroupKeyRequired   = errors.New("muscle group i18n key is required")
	ErrInvalidDifficulty        = errors.New("invalid difficulty level")
	ErrNameKeyRequired          = errors.New("name i18n key is required")
	ErrDescriptionKeyRequired   = errors.New("description i18n key is required")
)
