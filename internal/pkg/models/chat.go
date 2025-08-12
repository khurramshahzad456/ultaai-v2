package models

type ChatRequest struct {
	Message   string `json:"message"`
	UserToken string `json:"-"`
}
