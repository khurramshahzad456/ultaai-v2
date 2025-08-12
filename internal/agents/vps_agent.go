package agents

import (
	"ultahost-ai-gateway/internal/ai"
	"ultahost-ai-gateway/internal/pkg/models"
)

func HandleVPS(req *models.ChatRequest, functionList []string) (string, error) {
	functionName, err := ai.ClassifyFunctionWithinAgent(req.Message, functionList)

	if err != nil {
		return "", err
	}

	switch functionName {
	case "checkuptime":
		return checkUptime(req)
	case "checkdiskspace":
		return checkDiskSpace(req)
	case "installwordpress":
		return installWordPress(req)
	default:
		return "I couldn't match your request to a known VPS function.", nil
	}
}
