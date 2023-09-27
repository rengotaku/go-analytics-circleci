package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"rengotaku/go-analytics-circleci/circleciAPI"
	"rengotaku/go-analytics-circleci/libs"
	"rengotaku/go-analytics-circleci/models"

	"github.com/grezar/go-circleci"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	_ "time/tzdata"
)

type APIType int64

const (
	projectSlug string = "gh/jccapital/fundoor"

	ListPipelines APIType = iota
	ListWorkflowsByPipelineId
	ListWorkflowJobs
	DetailJobs
	RetrieveBranch
	SequenceColumns
	RetrieveStatus
	Unknown
)

func compAPITypeToString(s *string) APIType {
	switch *s {
	case "ListPipelines":
		return ListPipelines
	case "ListWorkflowsByPipelineId":
		return ListWorkflowsByPipelineId
	case "ListWorkflowJobs":
		return ListWorkflowJobs
	case "DetailJobs":
		return DetailJobs
	case "RetrieveBranch":
		return RetrieveBranch
	case "RetrieveStatus":
		return RetrieveStatus
	case "SequenceColumns":
		return SequenceColumns
	}

	return Unknown
}

var (
	db             *gorm.DB
	requestApiType APIType
)

func init() {
	os.Setenv("TZ", "Asia/Tokyo")

	log.SetLevel(log.DebugLevel)

	flags()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	envk := []string{"CIRCLE_TOKEN", "ORG_SLUG", "TARGET_ORG_SLUG", "RETRIEVED_TIME", "NEXT_PAGE_TOKEN"}
	log.Debugln("===Env")
	for _, k := range envk {
		log.Debugln(k, ": ", os.Getenv(k))
	}
	log.Debugln("===================")

	models.InitMigration()
	db, err = models.Connection()
	if err != nil {
		log.Fatal("DB Error: ", err)
	}
}

func flags() {
	a := flag.String("api", "ListPipelines", "Choose one of ListPipelines, ListWorkflowsByPipelineId, ListWorkflowJobs, DetailJobs, RetrieveBranch or SequenceColumns")
	flag.Parse()

	requestApiType = compAPITypeToString(a)
}

func main() {
	config := &circleci.Config{
		Token: os.Getenv("CIRCLE_TOKEN"),
	}

	client, err := circleci.NewClient(config)
	if err != nil {
		log.Fatal("Client:", err)
	}
	log.Debug("client: ", client)

	switch requestApiType {
	case ListPipelines:
		log.Debugln("ListPipelines executed")

		tk := os.Getenv("NEXT_PAGE_TOKEN")
		var nextPageToken *string
		if len(tk) > 0 {
			nextPageToken = &tk
		}

		lp := circleciAPI.NewListPipelines(client, projectSlug, db)
		lp.Execute(nextPageToken)
		return

	case ListWorkflowsByPipelineId:
		log.Debugln("ListWorkflowsByPipelineId executed")

		lw := circleciAPI.NewListWorkflowsByPipelineId(client, projectSlug, db)
		lw.Execute()

		return

	case ListWorkflowJobs:
		log.Debugln("ListWorkflowJobs executed")

		lwj := circleciAPI.NewListWorkflowJobs(client, projectSlug, db)
		lwj.Execute()

		return

	case DetailJobs:
		log.Debugln("DetailJobs executed")

		ldj := circleciAPI.NewListDetailJobs(client, projectSlug, db)
		ldj.Execute()

		return

	case RetrieveBranch:
		log.Debugln("RetrieveBranch executed")

		cnt, err := models.GetPipelineWitouhtBranchCount(db)
		if err != nil {
			log.Fatal("DB error: ", err)
		}

		oneTime := 10

		offset := int(cnt) / oneTime
		if int(cnt)%oneTime > 0 {
			offset += oneTime
		}

		log.Debug("offset: ", offset)

		for i := 0; i < offset; i++ {
			pipelines, err := models.GetAllWitouhtBranch(db, oneTime)
			if err != nil {
				log.Fatal("DB error: ", err)
			}

			for _, pipeline := range pipelines {
				var pip circleci.Pipeline
				if err := json.Unmarshal([]byte(pipeline.JSON), &pip); err != nil {
					panic(err)
				}

				pipeline.Branch = pip.Vcs.Branch
				db.Save(&pipeline)
			}
		}

		return

	case RetrieveStatus:
		log.Debugln("RetrieveStatus executed")

		cnt, err := models.GetWorkflowWitouhtStatus(db)
		if err != nil {
			log.Fatal("DB error: ", err)
		}

		oneTime := 10

		offset := int(cnt) / oneTime
		if int(cnt)%oneTime > 0 {
			offset += oneTime
		}

		log.Debug("offset: ", offset)

		for i := 0; i < offset; i++ {
			ws, err := models.GetAllWitouhtStatus(db, oneTime)
			if err != nil {
				log.Fatal("DB error: ", err)
			}

			for _, w := range ws {
				var workf circleci.Workflow
				if err := json.Unmarshal([]byte(w.JSON), &workf); err != nil {
					panic(err)
				}

				w.Status = fmt.Sprintf("%v", workf.Status)
				db.Save(&w)
			}
		}

		return

	case SequenceColumns:
		log.Debugln("SequenceColumns executed")

		libs.PrintSequences(db)

		return
	}

}
