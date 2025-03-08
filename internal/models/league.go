package models

// Model of the league table in the database
// Each record in this table covers a single league for a given season
type League struct {
	ID       uint16 // unique ID
	Division string // division of the league i.e. Open, IM, Pro
	SeasonID string // FK -> Season.ID
}
