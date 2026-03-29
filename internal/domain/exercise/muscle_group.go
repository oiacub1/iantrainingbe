package exercise

type MuscleGroup struct {
	Group             string `json:"group"`
	GroupKey          string `json:"groupKey"`
	ImpactPercentage  int    `json:"impactPercentage"`
}

const (
	MuscleGroupQuadriceps  = "QUADRICEPS"
	MuscleGroupHamstrings  = "HAMSTRINGS"
	MuscleGroupGlutes      = "GLUTES"
	MuscleGroupCalves      = "CALVES"
	MuscleGroupChest       = "CHEST"
	MuscleGroupBack        = "BACK"
	MuscleGroupShoulders   = "SHOULDERS"
	MuscleGroupBiceps      = "BICEPS"
	MuscleGroupTriceps     = "TRICEPS"
	MuscleGroupForearms    = "FOREARMS"
	MuscleGroupCore        = "CORE"
	MuscleGroupAbs         = "ABS"
	MuscleGroupObliques    = "OBLIQUES"
	MuscleGroupLowerBack   = "LOWER_BACK"
	MuscleGroupTraps       = "TRAPS"
)

func (mg MuscleGroup) Validate() error {
	if mg.ImpactPercentage < 0 || mg.ImpactPercentage > 100 {
		return ErrInvalidImpactPercentage
	}
	if mg.Group == "" {
		return ErrMuscleGroupRequired
	}
	if mg.GroupKey == "" {
		return ErrMuscleGroupKeyRequired
	}
	return nil
}

func ValidateMuscleGroups(groups []MuscleGroup) error {
	if len(groups) == 0 {
		return ErrNoMuscleGroups
	}
	
	totalImpact := 0
	for _, mg := range groups {
		if err := mg.Validate(); err != nil {
			return err
		}
		totalImpact += mg.ImpactPercentage
	}
	
	if totalImpact != 100 {
		return ErrInvalidTotalImpact
	}
	
	return nil
}
