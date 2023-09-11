package circleciAPI

import (
	"context"
	"encoding/json"
	"rengotaku/go-analytics-circleci/models"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/grezar/go-circleci"
	"gorm.io/gorm"
)

const (
	retrieveWorkflowPerOneTime = 10
)

type ListWorkflowJobs struct {
	client      *circleci.Client
	projectSlug string
	db          *gorm.DB
}

func NewListWorkflowJobs(client *circleci.Client, projectSlug string, db *gorm.DB) *ListWorkflowJobs {
	p := ListWorkflowJobs{client: client, projectSlug: projectSlug, db: db}
	return &p
}

func (l *ListWorkflowJobs) persistWorkflowJobs(wf models.Workflow) {
	log.Debug("WorkflowID: ", wf.WorkflowID)

	workflowJobList, err := l.client.Workflows.ListWorkflowJobs(context.Background(), wf.WorkflowID)
	if err != nil {
		log.Fatal("Got error as fetched workflowJobs: ", err)
	}
	wJobs := workflowJobList.Items
	log.Debug("workflowJobs: ", len(wJobs))

	for _, wj := range wJobs {
		jstr, _ := json.Marshal(*wj)

		w := models.WorkflowJob{WorkflowID: wf.WorkflowID, WorkflowJobID: wj.ID, JSON: string(jstr), JobNumber: wj.JobNumber, Status: wj.Status, Type: wj.Type, StartedAt: wj.StartedAt, StoppedAt: wj.StoppedAt}
		result := l.db.Create(&w)
		if result.Error != nil {
			log.Debug("Can't create workflowJob: ", result.Error)
			continue
		}
	}

	log.Debugln("sleeping...")
	time.Sleep(sleepTime)
}

func (l *ListWorkflowJobs) Execute() {
	cnt, err := models.GetWorkflowWitouhtJobCount(l.db)
	if err != nil {
		log.Fatal("DB error: ", err)
	}
	log.Debug("Target workflow's count: ", cnt)

	offset := int(cnt / retrieveWorkflowPerOneTime)
	if cnt%retrieveWorkflowPerOneTime > 0 {
		offset += retrieveWorkflowPerOneTime
	}

	log.Debug("offset: ", offset)

	for i := 0; i < offset; i++ {
		log.Debug("retrive count: ", i)
		workflows, err := models.GetAllWitouhtJob(l.db, retrieveWorkflowPerOneTime)
		if err != nil {
			log.Fatal("Got error as retrived workflows: ", err)
		}

		log.Debug("len(workflows): ", len(workflows))

		for _, wf := range workflows {
			l.persistWorkflowJobs(wf)
		}
	}
}
