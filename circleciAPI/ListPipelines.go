package circleciAPI

import (
	"context"
	"encoding/json"
	"os"
	"rengotaku/go-analytics-circleci/models"
	"time"

	"github.com/grezar/go-circleci"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	sleepTime time.Duration = 3 * time.Second
)

type ListPipelines struct {
	client      *circleci.Client
	projectSlug string
	db          *gorm.DB
}

func NewListPipelines(client *circleci.Client, projectSlug string, db *gorm.DB) *ListPipelines {
	p := ListPipelines{client: client, projectSlug: projectSlug, db: db}
	return &p
}

func (l *ListPipelines) FetchPipelineItems(nptoken *string, retrievedTime time.Time) {
	pipelines, err := l.client.Pipelines.List(context.Background(), circleci.PipelineListOptions{
		OrgSlug:   circleci.String(os.Getenv("ORG_SLUG")),
		PageToken: nptoken,
	})
	if err != nil {
		log.Fatal("List:", err)
	}

	log.Debugln("pipelines.Items count: ", len(pipelines.Items))

	overDate := false
	for _, pipeline := range pipelines.Items {
		if pipeline.ProjectSlug != l.projectSlug {
			log.Debugln("Does not intentd ProjectSlug")
			continue
		}
		if pipeline.CreatedAt.Before(retrievedTime) {
			log.Debugln("Retrieved Time: ", pipeline.CreatedAt)
			overDate = true
			continue
		}

		log.Debugln("PipelineID: ", pipeline.ID)
		jstr, _ := json.Marshal(*pipeline)

		var count int64
		l.db.Model(&models.Pipeline{}).Where("pipeline_id = ?", pipeline.ID).Count(&count)
		if count > 0 {
			log.Debugln("Existed ID: ", pipeline.ID)
			continue
		}

		pipeline := models.Pipeline{PipelineID: pipeline.ID, JSON: string(jstr), CreatedAt: pipeline.CreatedAt, Branch: pipeline.Vcs.Branch}
		result := l.db.Create(&pipeline)
		if result.Error != nil {
			log.Debug("Can't create pipeline: ", result.Error)
			continue
		}
	}

	if overDate {
		log.Debugln("Over the date.")
		return
	}

	log.Debugln("sleeping...")
	time.Sleep(sleepTime)

	if pipelines.NextPageToken == "" {
		return
	}

	log.Debug("Next page token: ", pipelines.NextPageToken)
	l.FetchPipelineItems(&pipelines.NextPageToken, retrievedTime)
}

func (l *ListPipelines) RetrivedTime(nextPageToken *string) time.Time {
	dateString := os.Getenv("RETRIEVED_TIME")
	rTime, err := time.Parse(time.RFC3339, dateString)

	if err != nil {
		log.Fatal("Time parsered error: ", err)
	}

	var count int64
	l.db.Model(&models.Pipeline{}).Count(&count)
	if count == 0 || nextPageToken != nil {
		return rTime
	}

	var pipeline *models.Pipeline
	l.db.Order("created_at desc").First(&pipeline)

	log.Debug(pipeline)
	if pipeline == nil {
		return rTime
	}

	if pipeline.CreatedAt.After(rTime) {
		return pipeline.CreatedAt
	} else {
		return rTime
	}
}

func (l *ListPipelines) Execute(nextPageToken *string) {
	t := l.RetrivedTime(nextPageToken)
	log.Debug("retrivedTime: ", t)

	l.FetchPipelineItems(nextPageToken, t)
}
