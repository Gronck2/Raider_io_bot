package main

type Gear struct {
	ItemLevelEquiped int `json:"item_level_equipped"`
	ItemLevelTotal   int `json:"item_level_total"`
	ArtifactTraits   int `json:"artifact_traits"`
}

type Guild struct {
	Name string `json:"name"`
}

type Season struct {
	Scores Scores `json:"scores"`
}

type Scores struct {
	All    float32 `json:"all"`
	Dps    float32 `json:"dps"`
	Healer float32 `json:"healer"`
	Tank   float32 `json:"tank"`
}

type Info struct {
	Name       string    `json:"name"`
	Race       string    `json:"race"`
	Class      string    `json:"class"`
	ProfileURL string    `json:"profile_url"`
	Gear       Gear      `json:"gear"`
	Guild      Guild     `json:"guild"`
	Season     []*Season `json:"mythic_plus_scores_by_season"`
}

type Affixes struct {
	Region         string          `json:"region"`
	Title          string          `json:"title"`
	LeaderboardURL string          `json:"leaderboard_url"`
	AffixDetails   []*AffixDetails `json:"affix_details"`
}

type AffixDetails struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WowheadURL  string `json:"wowhead_url"`
}

type ApiError struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}
