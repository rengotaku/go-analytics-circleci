package models

import "time"

type Job struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement:true"`
	WebURL           string `gorm:"type:text;"`
	JSON             string `gorm:"type:text;not null;"`
	Number           uint   `gorm:"not null;"`
	Name             string `gorm:"type:text;not null;"`
	Parallelism      int
	Status           string
	LatestWorkflowID string
	Duration         uint
	QueuedAt         time.Time
	StartedAt        time.Time
	StoppedAt        time.Time
	CreatedAt        time.Time
}
