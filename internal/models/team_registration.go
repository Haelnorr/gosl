package models

// Model of the team_registration table in the database
// Each row represents a teams application to play in a season
type TeamRegistration struct {
	ID              uint16 // unique ID
	TeamID          uint16 // FK -> Team.ID
	SeasonID        string // FK -> Season.ID
	PreferredLeague string // League the team prefers to play in
	Approved        uint16 // nil for pending, 0 for denied, 1 for accepted
	Placed          bool   // has the team been placed into a league?
}
