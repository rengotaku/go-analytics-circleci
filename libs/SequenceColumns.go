package libs

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SequenceResult struct {
	PipelineId string
	WorkflowId string
	Duration   int
	Status     string
	CreatedAt  string
	Seconds    int
}

func PrintSequences(db *gorm.DB) {
	var res []SequenceResult

	sql := `
		select 
			p.pipeline_id as pipeline_id, 
			w.workflow_id, 
			j.duration, 
			j.status, 
			datetime(j.created_at, '+9 hours') as created_at
		from 
			pipelines as p 
			inner join workflows as w on w.pipeline_id = p.pipeline_id 
			inner join workflow_jobs as wj on wj.workflow_id = w.workflow_id 
			inner join jobs as j on j.number = wj.job_number 
		where 
			p.branch = 'master' 
			and j.name = 'rspec' 
			and j.status in ('success', 'failed') 
		order by 
			j.created_at;
	`

	db.Raw(sql).Scan(&res)

	msr := make(map[string]SequenceResult)
	for _, r := range res {
		date, _ := time.Parse("2006-01-02 15:04:05", r.CreatedAt)
		r.Seconds = date.Second()

		if _, ok := msr[r.PipelineId]; !ok {
			msr[r.PipelineId] = r
		}
	}

	msr2 := make(map[string]SequenceResult)
	for _, v := range msr {
		msr2[strconv.Itoa(v.Seconds)] = v
	}
	keys := make([]string, 0, len(msr))
	for k, _ := range msr2 {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	msr3 := make(map[string]SequenceResult)
	for _, k := range keys {
		v := msr2[k]

		date, _ := time.Parse("2006-01-02 15:04:05", v.CreatedAt)
		if _, ok := msr3[date.Format("2006-01-02")]; !ok {
			msr3[date.Format("2006-01-02")] = v
		}
	}

	log.Debug("=======================================")
	for _, v := range msr3 {
		// 1000ms/60s = 1 minutes
		fmt.Printf("%v,%v,%v,%d,%v\n", v.PipelineId, v.WorkflowId, v.Status, v.Duration/1000/60, v.CreatedAt)
	}
	log.Debug("=======================================")

}
