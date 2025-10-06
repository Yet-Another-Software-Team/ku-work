package ai

import "ku-work/backend/model"

type ApprovalAI interface {
	CheckJob(job *model.Job) (model.JobApprovalStatus, []string)
	CheckStudent(student *model.Student) (model.StudentApprovalStatus, []string)
}
