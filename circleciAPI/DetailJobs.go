package circleciAPI

import (
	"context"
	"encoding/json"
	"fmt"
	"rengotaku/go-analytics-circleci/models"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/grezar/go-circleci"
	"gorm.io/gorm"
)

const (
	retrieveWorkflowJobDetailPerOneTime = 10
)

type DetailJobs struct {
	client      *circleci.Client
	projectSlug string
	db          *gorm.DB
}

func NewListDetailJobs(client *circleci.Client, projectSlug string, db *gorm.DB) *DetailJobs {
	p := DetailJobs{client: client, projectSlug: projectSlug, db: db}
	return &p
}

func (d *DetailJobs) deleteWorkflowJob(wfj models.WorkflowJob) {
	d.db.Where("workflow_job_id = ?", wfj.WorkflowJobID).Delete(&wfj)
	log.Debug("Deleted workflow: ", wfj.WorkflowJobID)
}

func (d *DetailJobs) persistWorkflowJobDetail(wfj models.WorkflowJob) {
	log.Debug("JobNumber: ", wfj.JobNumber)

	job, err := d.client.Jobs.Get(context.Background(), d.projectSlug, strconv.FormatInt(wfj.JobNumber, 10))
	if err != nil {
		log.Debug("Got error as fetched JobDetail: ", err)
		d.deleteWorkflowJob(wfj)
		return
	}

	jstr, _ := json.Marshal(*job)

	j := models.Job{
		WebURL:           job.WebURL,
		JSON:             string(jstr),
		Number:           uint(job.Number),
		Name:             job.Name,
		Parallelism:      job.Parallelism,
		Status:           fmt.Sprintf("%v", job.Status),
		LatestWorkflowID: job.LatestWorkflow.ID,
		Duration:         uint(job.Duration),
		QueuedAt:         job.QueuedAt,
		StartedAt:        job.StartedAt,
		StoppedAt:        job.StoppedAt,
		CreatedAt:        job.CreatedAt,
	}
	result := d.db.Create(&j)
	if result.Error != nil {
		log.Fatal("Can't create workflow: ", result.Error)
	}
}

func (d *DetailJobs) Execute() {
	log.Debug("DetailJobs exuecte")

	cnt, err := models.GetWorkflowJobWitouhtJobCount(d.db)
	if err != nil {
		log.Fatal("DB error: ", err)
	}
	log.Debug("Target workflowJob's count: ", cnt)

	offset := int(cnt / retrieveWorkflowJobDetailPerOneTime)
	if cnt%retrieveWorkflowJobDetailPerOneTime > 0 {
		offset += retrieveWorkflowJobDetailPerOneTime
	}

	log.Debug("offset: ", offset)

	for i := 0; i < offset; i++ {
		log.Debug("retrive count: ", i)
		wfjs, err := models.GetAllWitouhtWorkflowJob(d.db, retrieveWorkflowJobDetailPerOneTime)
		if err != nil {
			log.Fatal("Got error as retrived workflowJobs: ", err)
		}

		log.Debug("len(wfjs): ", len(wfjs))

		for _, wfj := range wfjs {
			if wfj.JobNumber == 0 {
				d.deleteWorkflowJob(wfj)
				continue
			}

			d.persistWorkflowJobDetail(wfj)

			log.Debugln("sleeping...")
			time.Sleep(sleepTime)
		}
	}
}
