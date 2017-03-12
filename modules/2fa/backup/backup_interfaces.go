package backup

import (
	"time"
)

// Code backup code instance interface
// Storer backup code objects must implement this interface
type Code interface {
	// Get user friendly name
	GetName() string
	// Get hashed secret
	GetHashedSecret() string
	// Check if the token has been used
	IsUsed() bool
	// Set a token used flag
	SetUsed()
	// Fetch the time a token was used
	GetUsed() time.Time
}

// Storer Backup Code store interface
// This must be implemented by a storage module to provide persistence to the module
type Storer interface {
	// Fetch a user instance by user id (should be able to remove this)
	GetUserByExtID(userid string) (interface{}, error)
	// Add a backup code to a given user
	AddBackupCode(userid, name, secret string) (interface{}, error)
	// Fetch backup codes for a given user
	GetBackupCodes(userid string) ([]interface{}, error)
	// Update a provided backup code
	UpdateBackupCode(code interface{}) (interface{}, error)
}
