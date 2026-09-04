package main

import (
	"encoding/json"
	"os"
)

type request struct {
	Protocol  int    `json:"protocol"`
	Operation string `json:"operation"`
	Request   struct {
		Inputs []string `json:"inputs"`
	} `json:"request"`
}

func main() {
	var input request
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": map[string]any{"code": "invalid_input", "message": err.Error()}})
		return
	}
	if len(input.Request.Inputs) == 0 {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": map[string]any{"code": "invalid_input", "message": "input is required"}})
		return
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"result": map[string]any{"operation": input.Operation, "data": map[string]any{"text": input.Request.Inputs[0]}}})
}
