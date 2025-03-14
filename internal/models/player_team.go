package models

import "time"

// Model of the player_team table in the database
// Each row represents a continuous period the player was on a team
type PlayerTeam struct {
	PlayerID uint16    // FK -> Player.ID
	TeamID   uint16    // FK -> Team.ID
	Start    time.Time // timestamp when player joined the team
	End      time.Time // timestamp when player left the team
}
