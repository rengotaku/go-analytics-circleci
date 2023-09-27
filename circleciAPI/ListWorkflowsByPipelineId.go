package circleciAPI

import (
	"context"
	"encoding/json"
	"fmt"
	"rengotaku/go-analytics-circleci/models"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/grezar/go-circleci"
	"gorm.io/gorm"
)

const (
	retrievePipelinePerOneTime = 50
)

type ListWorkflowsByPipelineId struct {
	client      *circleci.Client
	projectSlug string
	db          *gorm.DB
}

func NewListWorkflowsByPipelineId(client *circleci.Client, projectSlug string, db *gorm.DB) *ListWorkflowsByPipelineId {
	p := ListWorkflowsByPipelineId{client: client, projectSlug: projectSlug, db: db}
	return &p
}

func (l *ListWorkflowsByPipelineId) persistWorkflow(pipelines []models.Pipeline) {
	for _, pip := range pipelines {
		log.Debug("PipelineID: ", pip.PipelineID)

		workflowList, err := l.client.Pipelines.ListWorkflows(context.Background(), pip.PipelineID, circleci.PipelineListWorkflowsOptions{})
		if err != nil {
			log.Debug("Got error as fetched workflows: ", err)
			l.db.Where("pipeline_id = ?", pip.PipelineID).Delete(&pip)
			log.Debug("Deleted piplines: ", pip.PipelineID)
			continue
		}
		workflows := workflowList.Items
		pageToken := workflowList.NextPageToken

		for pageToken != "" {
			log.Debug("NextPageToken: ", pageToken)

			log.Debugln("sleeping...")
			time.Sleep(sleepTime)

			wl, err := l.client.Pipelines.ListWorkflows(context.Background(), pip.PipelineID, circleci.PipelineListWorkflowsOptions{
				PageToken: &pageToken,
			})
			if err != nil {
				log.Fatal("Got error as fetched workflows: ", err)
			}
			workflows = append(workflows, workflowList.Items...)
			pageToken = wl.NextPageToken
		}
		log.Debug("workflows: ", len(workflows))

		if len(workflows) == 0 {
			l.db.Where("pipeline_id = ?", pip.PipelineID).Delete(&pip)
			log.Debug("Deleted piplines: ", pip.PipelineID)
		}

		for _, workflow := range workflows {
			jstr, _ := json.Marshal(*workflow)

			w := models.Workflow{PipelineID: workflow.PipelineID, WorkflowID: workflow.ID, PipelineNumber: workflow.PipelineNumber, JSON: string(jstr), CreatedAt: workflow.CreatedAt, Status: fmt.Sprintf("%v", workflow.Status)}
			result := l.db.Create(&w)
			if result.Error != nil {
				log.Debug("Can't create workflow: ", result.Error)
				continue
			}
		}

		log.Debugln("sleeping...")
		time.Sleep(sleepTime)
	}
}

func (l *ListWorkflowsByPipelineId) Execute() {
	cnt, err := models.GetPipelineWitouhtWorkflowCount(l.db)
	if err != nil {
		log.Fatal("DB error: ", err)
	}

	offset := int(cnt / retrievePipelinePerOneTime)
	if cnt%retrievePipelinePerOneTime > 0 {
		offset += retrievePipelinePerOneTime
	}

	log.Debug("offset: ", offset)

	for i := 0; i < offset; i++ {
		log.Debug("retrive count: ", i)
		pipelines, err := models.GetAllWitouhtWorkflow(l.db, retrievePipelinePerOneTime)
		if err != nil {
			log.Fatal("Got error as retrived pipelines: ", err)
		}

		log.Debug("len(pipelines): ", len(pipelines))

		l.persistWorkflow(pipelines)
	}
}
