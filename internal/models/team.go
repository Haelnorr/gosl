package models

// Model of the team table in the database
// Each row represents a single team
type Team struct {
	ID           uint16 // unique ID
	Abbreviation string // unique abbreviation of team name
	Name         string // unique team name
	ManagerID    uint16 // FK -> Player.ID
}
