package models

import (
	"time"

	"gorm.io/gorm"
)

type Workflow struct {
	PipelineID string
	Pipeline   Pipeline

	WorkflowJobs []WorkflowJob

	WorkflowID     string `gorm:"primaryKey;type:text;"`
	JSON           string `gorm:"type:text;"`
	PipelineNumber int64
	CreatedAt      time.Time
	Status         string
}

func GetWorkflowWitouhtStatus(db *gorm.DB) (int64, error) {
	var cnt int64
	err := db.Model(&Workflow{}).Where("status is NULL or status = \"\"").Count(&cnt).Error
	return cnt, err
}

func GetAllWitouhtStatus(db *gorm.DB, limit int) ([]Workflow, error) {
	var ws []Workflow
	err := db.Model(&Workflow{}).Where("status is NULL or status = \"\"").Limit(limit).Find(&ws).Error
	return ws, err
}

func GetWorkflowWitouhtJobCount(db *gorm.DB) (int64, error) {
	var cnt int64
	err := db.Model(&Workflow{}).Joins("left join workflow_jobs on workflow_jobs.workflow_id = workflows.workflow_id").Where("workflow_jobs.workflow_id is NULL").Count(&cnt).Error
	return cnt, err
}

func GetAllWitouhtJob(db *gorm.DB, limit int) ([]Workflow, error) {
	var workflows []Workflow
	err := db.Model(&Workflow{}).Joins("left join workflow_jobs on workflow_jobs.workflow_id = workflows.workflow_id").Where("workflow_jobs.workflow_id is NULL").Limit(limit).Find(&workflows).Error
	return workflows, err
}
