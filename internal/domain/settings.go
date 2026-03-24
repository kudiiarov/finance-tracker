package domain

import "time"

// Settings represents user preferences.
type Settings struct {
	UserID       string    `json:"user_id"`
	MainCurrency string    `json:"main_currency"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DefaultSettings returns default settings for a new user.
func DefaultSettings(userID string) *Settings {
	now := time.Now()
	return &Settings{
		UserID:       userID,
		MainCurrency: "USD",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
