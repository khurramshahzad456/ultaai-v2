package models

type RunCommandRequest struct {
	Query    string `json:"query" binding:"required"`
	ServerIP string `json:"server_ip" binding:"required"`
	Hash     string `json:"hash" binding:"required"`
	UserID   string `json:"user_id,omitempty"`
}

type RunCommandResponse struct {
	Output      string `json:"output"`
	CommandUsed string `json:"command_used"`
	Explanation string `json:"explanation"`
}

type AICommand struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
}

type Step struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation,omitempty"`
}

type StepExecutionResponse struct {
	Description string       `json:"description"`
	UserID      string       `json:"user_id"`
	Steps       []StepResult `json:"steps"`
}

type StepResult struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation,omitempty"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
}

type StructuredAIResponse struct {
	Description string `json:"description"`
	Steps       []Step `json:"steps"`
}
