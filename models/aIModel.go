package models

// Calendar Model
type AIModel struct {
	Prompt string `json:"prompt" validate:"required"`
}
