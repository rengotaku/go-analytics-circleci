package models

import (
	"time"

	"gorm.io/gorm"
)

type Pipeline struct {
	PipelineID string `gorm:"primaryKey;type:text;not null;"`
	JSON       string `gorm:"type:text;not null;"`
	CreatedAt  time.Time
	Branch     string `gorm:"type:text;not null;"`

	Workflows []Workflow
}

func GetPipelineWitouhtBranchCount(db *gorm.DB) (int64, error) {
	var cnt int64
	err := db.Model(&Pipeline{}).Where("branch is NULL").Count(&cnt).Error
	return cnt, err
}

func GetAllWitouhtBranch(db *gorm.DB, limit int) ([]Pipeline, error) {
	var pipelines []Pipeline
	err := db.Model(&Pipeline{}).Where("branch is NULL").Limit(limit).Find(&pipelines).Error
	return pipelines, err
}

func GetPipelineWitouhtWorkflowCount(db *gorm.DB) (int64, error) {
	var cnt int64
	err := db.Model(&Pipeline{}).Joins("left join workflows on workflows.pipeline_id = pipelines.pipeline_id").Where("workflows.pipeline_id is NULL").Where("pipelines.branch = \"master\"").Count(&cnt).Error
	return cnt, err
}

func GetAllWitouhtWorkflow(db *gorm.DB, limit int) ([]Pipeline, error) {
	var pipelines []Pipeline
	err := db.Model(&Pipeline{}).Joins("left join workflows on workflows.pipeline_id = pipelines.pipeline_id").Where("workflows.pipeline_id is NULL").Where("pipelines.branch = \"master\"").Limit(limit).Find(&pipelines).Error
	return pipelines, err
}
