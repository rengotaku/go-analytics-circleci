package models

import (
	"time"

	"gorm.io/gorm"
)

type WorkflowJob struct {
	WorkflowID string `gorm:"primaryKey;type:text;"`
	Workflow   Workflow

	WorkflowJobID string    `gorm:"primaryKey;type:text;"`
	JSON          string    `gorm:"type:text;"`
	JobNumber     int64     `json:"job_number"`
	Status        string    `json:"status"`
	Type          string    `json:"type"`
	StartedAt     time.Time `json:"started_at"`
	StoppedAt     time.Time `json:"stopped_at"`
}

func GetWorkflowJobWitouhtJobCount(db *gorm.DB) (int64, error) {
	var cnt int64
	err := db.Model(&WorkflowJob{}).Joins("left join jobs on jobs.number = workflow_jobs.job_number").Where("jobs.number is NULL").Count(&cnt).Error
	return cnt, err
}

func GetAllWitouhtWorkflowJob(db *gorm.DB, limit int) ([]WorkflowJob, error) {
	var wjs []WorkflowJob
	err := db.Model(&WorkflowJob{}).Joins("left join jobs on jobs.number = workflow_jobs.job_number").Where("jobs.number is NULL").Limit(limit).Find(&wjs).Error
	return wjs, err
}
