package gamelogs

type Gamelog struct {
	Type            string `json:"type"`
	MatchID         string `json:"match_id"` // TODO: change to UUID
	Winner          string `json:"winner"`
	Arena           string `json:"arena"`
	PeriodsEnabled  string `json:"periods_enabled"`
	CurrentPeriod   string `json:"current_period"`
	CustomMercyRule string `json:"custom_mercy_rule"`
	MatchLength     string `json:"match_length"`
	EndReason       string `json:"end_reason"`
	Score           struct {
		Home uint16 `json:"home"`
		Away uint16 `json:"away"`
	} `json:"score"`
	Players []struct {
		GameUserID string `json:"game_user_id"`
		Team       string `json:"team"`
		Username   string `json:"username"`
		Stats      struct {
			PeriodsPlayed     float32 `json:"periods_played"`
			Passes            float32 `json:"passes"`
			Turnovers         float32 `json:"turnovers"`
			Takeaways         float32 `json:"takeaways"`
			ConcededGoals     float32 `json:"conceded_goals"`
			Blocks            float32 `json:"blocks"`
			Score             float32 `json:"score"`
			PossessionTimeSec float32 `json:"possession_time_sec"`
			Saves             float32 `json:"saves"`
			Assists           float32 `json:"assists"`
			PrimaryAssists    float32 `json:"primary_assists"`
			SecondaryAssists  float32 `json:"secondary_assists"`
			Goals             float32 `json:"goals"`
			ContributedGoals  float32 `json:"contributed_goals"`
			Shots             float32 `json:"shots"`
			PostHits          float32 `json:"post_hits"`
			FaceoffsWon       float32 `json:"faceoffs_won"`
			FaceoffsLost      float32 `json:"faceoffs_lost"`
			GameWinningGoals  float32 `json:"game_winning_goals"`
			Wins              float32 `json:"wins"`
			Losses            float32 `json:"losses"`
			OvertimeWins      float32 `json:"overtime_wins"`
			OvertimeGoals     float32 `json:"overtime_goals"`
			OvertimeLosses    float32 `json:"overtime_losses"`
		} `json:"stats"`
	} `json:"players"`
	// NOTE: these dont seem to be in the api, added by the log extractor
	// ask gurgur what they're for
	PreCopy struct {
		CR    float32 `json:"cr"`
		Mod   float32 `json:"mod"`
		Check uint16  `json:"check"`
	} `json:"preCopy"`
	Copy float32 `json:"copy"`
}
