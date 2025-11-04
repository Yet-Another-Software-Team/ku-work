package ai

import (
	"ku-work/backend/model"
)

type DummyApprovalAI struct {
}

func NewDummyApprovalAI() *DummyApprovalAI {
	return &DummyApprovalAI{}
}

func (current *DummyApprovalAI) CheckJob(job *model.Job) (model.JobApprovalStatus, []string) {
	return model.JobApprovalPending, nil
}

func (current *DummyApprovalAI) CheckStudent(student *model.Student) (model.StudentApprovalStatus, []string) {
	return model.StudentApprovalPending, nil
}
