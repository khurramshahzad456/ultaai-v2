package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ultahost-ai-assistant/internal/ai"
	"ultahost-ai-assistant/internal/executor"
	"ultahost-ai-assistant/pkg/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Executor *executor.HTTPExecutor
}

func NewHandler() *Handler {
	return &Handler{
		Executor: executor.NewHTTPExecutor(5 * time.Second),
	}
}

func (h *Handler) RunCommand(c *gin.Context) {
	var req models.RunCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format", "details": err.Error()})
		return
	}

	// Try multi-step first
	steps, err := ai.ExtractStepsFromPrompt(req.Query)
	if err == nil && len(steps.Steps) > 0 {
		var results []models.StepResult


		vmURL := fmt.Sprintf("http://%s:8080/run-command", req.ServerIP)

		for _, step := range steps.Steps {
			if !strings.HasPrefix(step.Command, "sudo") {
				results = append(results, models.StepResult{Command: step.Command, Error: "Command rejected by policy"})
				continue
			}

			body, _ := json.Marshal(map[string]string{"command": step.Command, "hash": req.Hash})
			resp, err := http.Post(vmURL, "application/json", bytes.NewBuffer(body))
			result := models.StepResult{Command: step.Command, Explanation: step.Explanation}

			if err != nil {
				result.Error = fmt.Sprintf("Request failed: %v", err)
				results = append(results, result)
				break
			}

			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK {
				result.Error = string(b)
				results = append(results, result)
				break
			}
			result.Output = string(b)
			results = append(results, result)
		}

		c.JSON(http.StatusOK, models.StepExecutionResponse{
			Description: steps.Description,
			UserID:      req.UserID,
			Steps:       results,
		})
		return
	}

	// fallback: single command
	aiCmd, err := ai.ExtractCommandFromPrompt(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI failed", "details": err.Error()})
		return
	}

	if !executor.IsCommandAllowed(aiCmd.Command) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Command not allowed"})
		return
	}

	vmRes, err := h.Executor.Execute(req.ServerIP, aiCmd.Command, req.Hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "VM error", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.RunCommandResponse{
		Output:      vmRes.Output,
		CommandUsed: aiCmd.Command,
		Explanation: aiCmd.Explanation,
	})
}
