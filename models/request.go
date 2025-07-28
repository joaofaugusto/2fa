package models

type SendCodeRequest struct {
	Email string `json:"email"`
}

type SendCodeResponse struct {
	Message string `json:"message"`
}

type VerifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
