package handlers

import (
	"errors"
	"ku-work/backend/handlers/ai"
	"ku-work/backend/model"
	"os"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type AIHandler struct {
	DB *gorm.DB
	AI ai.ApprovalAI
}

func (current *AIHandler) AutoApproveJob(job *model.Job) {
	approvalStatus, reasons := current.AI.CheckJob(job)
	if approvalStatus == model.JobApprovalPending {
		return
	}
	tx := current.DB.Begin()
	if err := current.DB.Model(&model.Job{
		ID: job.ID,
	}).Update("approval_status", approvalStatus).Error; err != nil {
		tx.Rollback()
		return
	}
	if err := tx.Create(&model.Audit{
		ActorID:    "ai",
		Action:     string(approvalStatus),
		ObjectName: "Job",
		Reason:     "- " + strings.Join(reasons, "\n- "),
		ObjectID:   strconv.FormatUint(uint64(job.ID), 10),
	}).Error; err != nil {
		tx.Rollback()
		return
	}
	_ = tx.Commit()
}

func (current *AIHandler) AutoApproveStudent(student *model.Student) {
	// Use AI to check student status
	// This might take a while
	approvalStatus, reasons := current.AI.CheckStudent(student)

	// Maybe error occur so it returns
	if approvalStatus == model.StudentApprovalPending {
		return
	}

	// We refetch because since AI take time it might be stale now
	tx := current.DB.Begin()
	if err := current.DB.Model(&model.Student{
		UserID: student.UserID,
	}).Update("approval_status", approvalStatus).Error; err != nil {
		tx.Rollback()
		return
	}
	if err := tx.Create(&model.Audit{
		ActorID:    "ai",
		Action:     string(approvalStatus),
		ObjectName: "Student",
		Reason:     "- " + strings.Join(reasons, "\n- "),
		ObjectID:   student.UserID,
	}).Error; err != nil {
		tx.Rollback()
		return
	}
	_ = tx.Commit()
}

func NewAIHandler(DB *gorm.DB) (*AIHandler, error) {
	approvalAIName, hasApprovalAI := os.LookupEnv("APPROVAL_AI")
	if !hasApprovalAI {
		return nil, errors.New("approval ai not specified")
	}
	switch approvalAIName {
	case "ollama":
		approvalAI, err := ai.NewOllamaApprovalAI()
		if err != nil {
			return nil, err
		}
		return &AIHandler{
			DB: DB,
			AI: approvalAI,
		}, nil
	case "dummy":
		return &AIHandler{
			DB: DB,
			AI: ai.NewDummyApprovalAI(),
		}, nil
	}
	return nil, errors.New("invalid approval ai specified")
}
