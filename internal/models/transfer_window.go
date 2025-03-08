package models

import "time"

// Model of the transfer_window table in the database
type TransferWindow struct {
	ID       uint16    // unique ID
	SeasonID string    // FK -> Season.ID
	Start    time.Time // start of the transfer window
	End      time.Time // end of the transfer window
}
