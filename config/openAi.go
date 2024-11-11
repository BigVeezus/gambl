package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const apiURL = "https://api.openai.com/v1/chat/completions" // Endpoint URL

func AskOpenAI(message string) (map[string]interface{}, error) {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_KEY")

	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": "Translate the following English text to French: 'Hello, how are you?'"},
		},
		"max_tokens":  60,
		"temperature": 0.5,
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("Error marshaling request data: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Error unmarshaling response: %v", err)
		return nil, err
	}

	fmt.Println("Response:", result)
	return result, nil

}
