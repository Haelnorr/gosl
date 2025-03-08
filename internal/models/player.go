package models

// Model of the player table in the database
// Each row represents a single player
type Player struct {
	ID        uint16 // unique ID
	SlapID    uint32 // unique slapshot player ID
	Name      string // unique player name
	DiscordID string // unique discord ID
}
