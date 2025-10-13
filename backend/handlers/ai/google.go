package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"ku-work/backend/model"
	"log"
	"os"
	"path/filepath"
	"strings"

	"code.sajari.com/docconv"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type googleAI struct {
	db     *gorm.DB
	client *genai.GenerativeModel
}

func NewGoogleAI(db *gorm.DB) (*googleAI, error) {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_AI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-2.5-flash-lite") //TODO: externalize this
	model.GenerationConfig.ResponseMIMEType = "application/json"

	return &googleAI{db: db, client: model}, nil
}

type ResponseType struct {
	Valid   bool     `json:"valid"`
	Reasons []string `json:"reasons"`
}

func (g *googleAI) getCompletion(parts []genai.Part) (*ResponseType, error) {
	ctx := context.Background()

	resp, err := g.client.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	rawResponse := string(resp.Candidates[0].Content.Parts[0].(genai.Text))
	cleanResponse := strings.TrimSpace(rawResponse)

	if strings.HasPrefix(cleanResponse, "```json") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)
	} else if strings.HasPrefix(cleanResponse, "```") {
		cleanResponse = strings.TrimPrefix(cleanResponse, "```")
		cleanResponse = strings.TrimSuffix(cleanResponse, "```")
		cleanResponse = strings.TrimSpace(cleanResponse)

	}

	response := &ResponseType{}
	if err := json.Unmarshal([]byte(cleanResponse), response); err != nil {
		return nil, err
	}

	return response, nil
}

func (g *googleAI) CheckJob(job *model.Job) (model.JobApprovalStatus, []string) {
	if err := g.db.Preload("Company.User").First(job, job.ID).Error; err != nil {
		log.Printf("error preloading job data for job %d: %v", job.ID, err)
		return model.JobApprovalPending, []string{"Failed to load job data"}
	}

	type AIJobInput struct {
		Name        string               `json:"name,omitempty"`
		CompanyName string               `json:"companyName,omitempty"`
		Position    string               `json:"position,omitempty"`
		Duration    string               `json:"duration,omitempty"`
		Description string               `json:"description,omitempty"`
		Location    string               `json:"location,omitempty"`
		JobType     model.JobType        `json:"jobType,omitempty"`
		Experience  model.ExperienceType `json:"experienceType,omitempty"`
		MinSalary   uint                 `json:"minSalary,omitempty"`
		MaxSalary   uint                 `json:"maxSalary,omitempty"`
	}

	jobData, err := json.Marshal(AIJobInput{
		Name:        job.Name,
		CompanyName: job.Company.User.Username,
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
		return model.JobApprovalPending, []string{"Failed to marshal job data"}
	}

	prompt := "Please evaluate whether the following job posting is valid. A valid job should be professional and not spam, a scam, or inappropriate. Provide short, descriptive reasons for your evaluation. Salary is in Baht per month. Ignore missing contact information. Respond in JSON format specified in the schema. The json schema is: {\"type\":\"object\",\"properties\":{\"reasons\":{\"type\":\"array\"},\"valid\":{\"type\":\"boolean\"}}}. The job data is as follows: " + string(jobData)

	response, err := g.getCompletion([]genai.Part{genai.Text(prompt)})
	if err != nil {
		log.Printf("error getting completion: %v", err)
		return model.JobApprovalPending, []string{"Error getting completion from AI"}
	}

	if response.Valid {
		return model.JobApprovalAccepted, response.Reasons
	}

	return model.JobApprovalRejected, response.Reasons
}

func (g *googleAI) CheckStudent(student *model.Student) (model.StudentApprovalStatus, []string) {
	if err := g.db.Preload("StudentStatusFile").First(student, student.User.ID).Error; err != nil {
		log.Printf("error preloading student data for student %d: %v", student.User.ID, err)
		return model.StudentApprovalPending, []string{"Failed to load student data"}
	}

	studentData, err := json.Marshal(student)
	if err != nil {
		return model.StudentApprovalPending, []string{"Failed to marshal student data"}
	}

	prompt := `You are a document validation AI. Your task is to determine if the provided 'StudentStatusFile' is valid proof of student status at Kasetsart University.

**Validation Rules:**

1.  **University Check:** The document MUST explicitly mention "Kasetsart University". If it does not, or mentions another university, it is INVALID.
2.  **Document Type Check:** The document's content and structure must clearly match one of the following accepted types:
    *   Student ID Card
    *   Academic Transcript
    *   Certificate of Graduation
3.  **Content Analysis:** A research paper, a resume, a personal letter, or any other document type is NOT acceptable. You must analyze the text to determine the document's true nature.
4.  **Data Corroboration:** The information in the document (like name or student status) should align with the provided student data.

**Your Response:**

-   If the document is invalid, set "valid" to false and provide a reason.
-   If the document is a valid type but from the wrong university, the reason must state: "The document is from [Identified University], not Kasetsart University."
-   If the document is an unacceptable type (like a research paper), the reason must state: "The document appears to be a [Identified Document Type], which is not an acceptable proof of student status."

Respond with ONLY a raw JSON object, adhering to this schema: {"type":"object","properties":{"reasons":{"type":"array"},"valid":{"type":"boolean"}}}. Do not include Markdown formatting.

The student data is as follows: ` + string(studentData)
	parts := []genai.Part{genai.Text(prompt)}

	if student.StudentStatusFile.ID != "" {
		filePath := filepath.Join("./files", student.StudentStatusFile.ID)
		switch student.StudentStatusFile.Category {
			case model.FileCategoryImage:
				fileData, err := os.ReadFile(filePath)
				if err == nil {
					parts = append(parts, genai.ImageData(string(student.StudentStatusFile.FileType), fileData))
				}
			case model.FileCategoryDocument:
				text, err := docconv.ConvertPath(filePath)
				if err == nil {
					parts = append(parts, genai.Text("\n\nStudent Status File Content:\n"+text.Body))
				}
		}
	}

	response, err := g.getCompletion(parts)
	if err != nil {
		log.Printf("error getting completion: %v", err)
		return model.StudentApprovalPending, []string{"Error getting completion from AI"}
	}

	if response.Valid {
		return model.StudentApprovalAccepted, response.Reasons
	}

	return model.StudentApprovalRejected, response.Reasons
}
