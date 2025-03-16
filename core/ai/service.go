package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	gameCore "gambl/core/game"
)

type AIService interface {
	// Game Management
	ValidateWager(ctx context.Context, game *gameCore.Game) (*gameCore.ValidationResult, error)
	// buildValidationPrompt(game *gameCore.Game) (*gameCore.Game, error)
	ResolutionCheck(ctx context.Context, game *gameCore.Game) (*gameCore.ResolutionResult, error)
	// buildResolutionPrompt(game **gameCore.Game) string
}

type AIController struct {
	apiKey string
	apiURL string
}

func NewAIService(apiKey string) *AIController {
	return &AIController{
		apiKey: apiKey,
		apiURL: "https://api.deepseek.com/v1/chat/completions", // Replace with DeepSeek’s actual API endpoint
	}
}

// ValidateWager checks if a wager proposition is valid and resolvable
func (c *AIController) ValidateWager(ctx context.Context, wager *gameCore.Game) (*gameCore.ValidationResult, error) {
	prompt := buildValidationPrompt(wager)

	reqBody := map[string]interface{}{
		"model": "deepseek-chat", // Check DeepSeek docs for correct model name
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are a wager validator that evaluates betting propositions.
                Your task is to determine if a wager is valid and resolvable.
                Respond with a JSON object containing these fields:
                {"isValid": true/false, "reason": "explanation", "isResolvable": true/false, "resolutionGuidance": "explanation"}`,
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3, // Lower temperature for more consistent results
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second} // Increased timeout for AI responses
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call DeepSeek API: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("API error: %d - %v", resp.StatusCode, errorResp)
	}

	var deepSeekResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	// Parse response structure
	if err := json.NewDecoder(resp.Body).Decode(&deepSeekResp); err != nil {
		return nil, fmt.Errorf("failed to parse DeepSeek response: %w", err)
	}

	if len(deepSeekResp.Choices) == 0 {
		return nil, fmt.Errorf("AI response has no choices")
	}

	content := deepSeekResp.Choices[0].Message.Content

	// Try to parse as JSON, but have fallback handling
	var result gameCore.ValidationResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Fallback: create a basic result with the error
		log.Printf("Failed to parse AI response as JSON: %v", err)
		log.Printf("Raw response content: %s", content)
		return &gameCore.ValidationResult{
			IsValid:   false,
			Reasoning: "Failed to parse AI response",
		}, nil
	}

	return &result, nil
}

// buildValidationPrompt creates a detailed prompt from the wager data
func buildValidationPrompt(wager *gameCore.Game) string {
	// Format the date if available
	endDateStr := "not specified"
	if !wager.Deadline.IsZero() {
		endDateStr = wager.Deadline.Format("January 2, 2006")
	}

	prompt := fmt.Sprintf(`
Please evaluate this wager proposal:

Wager statement: "%s"
Proposed by: User #%s
Proposed resolution date: %s
Supporting details provided: "%s"

Determine if this wager is valid and resolvable based on the criteria in your instructions.
`, wager.Statement, wager.Creator_ID, endDateStr, wager.Description)

	return prompt
}

// ResolutionCheck attempts to resolve a mature wager
func (c *AIController) ResolutionCheck(ctx context.Context, wager *gameCore.Game) (*gameCore.ResolutionResult, error) {
	prompt := buildResolutionPrompt(wager)

	reqBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a wager resolution system...",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	reqBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call DeepSeek API: %w", err)
	}
	defer resp.Body.Close()

	var deepSeekResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&deepSeekResp); err != nil {
		return nil, fmt.Errorf("failed to parse DeepSeek response: %w", err)
	}

	var result gameCore.ResolutionResult
	if err := json.Unmarshal([]byte(deepSeekResp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &result, nil
}

// buildResolutionPrompt creates a detailed prompt for wager resolution
func buildResolutionPrompt(wager *gameCore.Game) string {
	currentDate := time.Now().Format("January 2, 2006")

	prompt := fmt.Sprintf(`
Please check if this wager has been resolved:

Wager statement: "%s"
Created on: %s
Resolution deadline: %s
Today's date: %s
Supporting details provided: "%s"

Has this wager been resolved? If so, did the "support" position or the "oppose" position win?
Provide your reasoning and any relevant evidence.
`,
		wager.Statement,
		wager.Created_At.Format("January 2, 2006"),
		wager.Deadline.Format("January 2, 2006"),
		currentDate,
		wager.Description)

	return prompt
}
