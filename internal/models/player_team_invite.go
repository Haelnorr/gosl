package models

// Model of the player_team_invite table in the database
// Each row represents an invite sent to a player to join a team
type PlayerTeamInvite struct {
	ID       uint32 // unique ID
	PlayerID uint16 // FK -> Player.ID
	TeamID   uint16 // FK -> Team.ID
	Status   uint16 // nil for pending, 0 for rejected, 1 for accepted
	Approved uint16 // nil for pending, 0 for denied, 1 for approved
}
