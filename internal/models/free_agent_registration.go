package models

// Model of the free_agent_registration table in the database
// Each row represents a request for a player to play as a free agent in a season
type FreeAgentRegistration struct {
	ID              uint32 // unique ID
	PlayerID        uint16 // FK -> Player.ID
	SeasonID        string // FK -> Season.ID
	PreferredLeague string // League the player prefers to play in
	Approved        uint16 // nil for pending, 0 for denied, 1 for accepted
	Placed          bool   // has the player been placed into a league?
}
