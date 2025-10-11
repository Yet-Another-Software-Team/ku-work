package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"ku-work/backend/model"
	"net/http"
	"net/url"
	"os"
)

type OllamaApprovalAI struct {
	model  string
	uri    *url.URL
	client http.Client
}

func NewOllamaApprovalAI() (*OllamaApprovalAI, error) {
	model, has_model := os.LookupEnv("APPROVAL_AI_MODEL")
	if !has_model {
		return nil, errors.New("approval ai model not specified")
	}
	uri, has_uri := os.LookupEnv("APPROVAL_AI_URI")
	if !has_uri {
		return nil, errors.New("approval ai uri not specified")
	}
	parsed_url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	return &OllamaApprovalAI{
		model: model,
		uri:   parsed_url,
	}, nil
}

type AIOptions struct {
	Model  string          `json:"model"`
	Prompt string          `json:"prompt"`
	Format json.RawMessage `json:"format"`
	System string          `json:"system"`
	Stream bool            `json:"stream"`
}

func (current *OllamaApprovalAI) CheckJob(job *model.Job) (model.JobApprovalStatus, []string) {
	type AIInput struct {
		Name        string               `json:"name,omitempty"`
		Position    string               `json:"position,omitempty"`
		Duration    string               `json:"duration,omitempty"`
		Description string               `json:"description,omitempty"`
		Location    string               `json:"location,omitempty"`
		JobType     model.JobType        `json:"jobType,omitempty"`
		Experience  model.ExperienceType `json:"experienceType,omitempty"`
		MinSalary   uint                 `json:"minSalary,omitempty"`
		MaxSalary   uint                 `json:"maxSalary,omitempty"`
	}
	jobData, err := json.Marshal(AIInput{
		Name:        job.Name,
		Position:    job.Position,
		Duration:    job.Duration,
		Description: job.Description,
		Location:    job.Location,
		JobType:     job.JobType,
		Experience:  job.Experience,
		MinSalary:   job.MinSalary,
		MaxSalary:   job.MaxSalary,
	})
	if err != nil {
		return model.JobApprovalPending, nil
	}
	optsData, err := json.Marshal(AIOptions{
		Model:  current.model,
		System: "Please evaluate whether job application is valid or not. Salary unit is Baht per month. Ignore missing company name, and contact information. Please respond in JSON",
		Prompt: string(jobData),
		Format: json.RawMessage(`{"type":"object","properties":{"reasons":{"type":"array"},"valid":{"type":"boolean"}}}`),
		Stream: false,
	})
	if err != nil {
		return model.JobApprovalPending, nil
	}
	resp, err := current.client.Post(current.uri.JoinPath("api", "generate").String(), "application/json", bytes.NewReader(optsData))
	if err != nil {
		return model.JobApprovalPending, nil
	}
	type OllamaResponse struct {
		Response string `json:"response"`
	}
	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.JobApprovalPending, nil
	}
	ollamaResponse := OllamaResponse{}
	if err := json.Unmarshal(rawResponse, &ollamaResponse); err != nil {
		return model.JobApprovalPending, nil
	}
	type ResponseType struct {
		Valid   bool     `json:"valid"`
		Reasons []string `json:"reasons"`
	}
	response := ResponseType{}
	if json.Unmarshal([]byte(ollamaResponse.Response), &response) != nil {
		return model.JobApprovalPending, nil
	}
	if response.Valid {
		return model.JobApprovalAccepted, response.Reasons
	}
	return model.JobApprovalRejected, response.Reasons
}

// This just checks fields only, not file.
// File comes in different format, docx, doc, pdf, csv, png, jpg.
// Maybe you can figure out how to convert all of those and input it in AI somehow.
func (current *OllamaApprovalAI) CheckStudent(student *model.Student) (model.StudentApprovalStatus, []string) {
	studentData, err := json.Marshal(&student)
	if err != nil {
		return model.StudentApprovalPending, nil
	}
	optsData, err := json.Marshal(AIOptions{
		Model:  current.model,
		System: "Please evaluate whether student is a real valid student or just some one trolling pretending to be one. Respond in JSON.",
		Prompt: string(studentData),
		Format: json.RawMessage(`{"type":"object","properties":{"reasons":{"type":"array"},"valid":{"type":"boolean"}}}`),
		Stream: false,
	})
	if err != nil {
		return model.StudentApprovalPending, nil
	}
	resp, err := current.client.Post(current.uri.JoinPath("api", "generate").String(), "application/json", bytes.NewReader(optsData))
	if err != nil {
		return model.StudentApprovalPending, nil
	}
	type OllamaResponse struct {
		Response string `json:"response"`
	}
	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.StudentApprovalPending, nil
	}
	ollamaResponse := OllamaResponse{}
	if err := json.Unmarshal(rawResponse, &ollamaResponse); err != nil {
		return model.StudentApprovalPending, nil
	}
	type ResponseType struct {
		Valid   bool     `json:"valid"`
		Reasons []string `json:"reasons"`
	}
	response := ResponseType{}
	if json.Unmarshal([]byte(ollamaResponse.Response), &response) != nil {
		return model.StudentApprovalPending, nil
	}
	if response.Valid {
		return model.StudentApprovalAccepted, response.Reasons
	}
	return model.StudentApprovalRejected, response.Reasons
}
