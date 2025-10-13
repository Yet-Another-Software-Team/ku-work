package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"ku-work/backend/handlers/ai"
	"ku-work/backend/model"
	"os"
	"strconv"
	"strings"
	"text/template"

	"gorm.io/gorm"
)

type AIHandler struct {
	DB                                       *gorm.DB
	AI                                       ai.ApprovalAI
	emailHandler                             *EmailHandler
	jobApprovalStatusUpdateEmailTemplate     *template.Template
	studentApprovalStatusUpdateEmailTemplate *template.Template
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
	reasonsString := "- " + strings.Join(reasons, "\n- ")
	if err := tx.Create(&model.Audit{
		ActorID:    "ai",
		Action:     string(approvalStatus),
		ObjectName: "Job",
		Reason:     reasonsString,
		ObjectID:   strconv.FormatUint(uint64(job.ID), 10),
	}).Error; err != nil {
		tx.Rollback()
		return
	}
	_ = tx.Commit()

	type Context struct {
		Company model.Company
		User    model.User
		Job     *model.Job
		Status  string
		Reason  string
	}
	var context Context
	context.Company.UserID = job.CompanyID
	if err := current.DB.Select("email").Take(&context.Company).Error; err != nil {
		return
	}
	context.User.ID = job.CompanyID
	if err := current.DB.Select("username").Take(&context.User).Error; err != nil {
		return
	}
	context.Job = job
	context.Status = string(job.ApprovalStatus)
	context.Reason = reasonsString
	var tpl bytes.Buffer
	if err := current.jobApprovalStatusUpdateEmailTemplate.Execute(&tpl, context); err != nil {
		return
	}
	_ = current.emailHandler.provider.SendTo(
		context.Company.Email,
		fmt.Sprintf("[KU-WORK] Your \"%s\" job has been automatically reviewed", job.Name),
		tpl.String(),
	)
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
	reasonsString := "- " + strings.Join(reasons, "\n- ")
	if err := tx.Create(&model.Audit{
		ActorID:    "ai",
		Action:     string(approvalStatus),
		ObjectName: "Student",
		Reason:     reasonsString,
		ObjectID:   student.UserID,
	}).Error; err != nil {
		tx.Rollback()
		return
	}
	_ = tx.Commit()

	type Context struct {
		OAuth  model.GoogleOAuthDetails
		Status string
		Reason string
	}
	var context Context
	context.OAuth.UserID = student.UserID
	context.Status = string(student.ApprovalStatus)
	context.Reason = reasonsString
	if err := current.DB.Select("email,first_name,last_name").Take(&context.OAuth).Error; err != nil {
		return
	}
	var tpl bytes.Buffer
	if err := current.studentApprovalStatusUpdateEmailTemplate.Execute(&tpl, context); err != nil {
		return
	}
	_ = current.emailHandler.provider.SendTo(
		context.OAuth.Email,
		"[KU-WORK] Your student account has been automatically reviewed",
		tpl.String(),
	)
}

func NewAIHandler(DB *gorm.DB, emailHandler *EmailHandler) (*AIHandler, error) {
	approvalAIName, hasApprovalAI := os.LookupEnv("APPROVAL_AI")
	if !hasApprovalAI {
		return nil, errors.New("approval ai not specified")
	}
	jobApprovalStatusUpdateEmailTemplate, err := template.New("job_auto_approval_status_update.tmpl").ParseFiles("email_templates/job_auto_approval_status_update.tmpl")
	if err != nil {
		return nil, err
	}
	studentApprovalStatusUpdateEmailTemplate, err := template.New("student_auto_approval_status_update.tmpl").ParseFiles("email_templates/student_auto_approval_status_update.tmpl")
	if err != nil {
		return nil, err
	}
	aiHandler := &AIHandler{
		DB:                                       DB,
		jobApprovalStatusUpdateEmailTemplate:     jobApprovalStatusUpdateEmailTemplate,
		studentApprovalStatusUpdateEmailTemplate: studentApprovalStatusUpdateEmailTemplate,
		emailHandler:                             emailHandler,
	}
	switch approvalAIName {
	case "ollama":
		approvalAI, err := ai.NewOllamaApprovalAI()
		if err != nil {
			return nil, err
		}
		aiHandler.AI = approvalAI
		return aiHandler, nil
	case "dummy":
		aiHandler.AI = ai.NewDummyApprovalAI()
		return aiHandler, nil
	}
	return nil, errors.New("invalid approval ai specified")
}
