package display

import (
	"encoding/json"
	"fmt"
)

type CLIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func PrintJSON(status, message string, data interface{}, err error) {
	resp := CLIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	bytess, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(bytess))
}
