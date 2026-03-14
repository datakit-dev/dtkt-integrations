package v1beta1

import (
	"encoding/base64"
	"encoding/json"
)

type queryPage struct {
	JobID     string     `json:"job_id"`
	PageToken string     `json:"page_token"`
	PrevPage  *queryPage `json:"prev_page,omitempty"`
}

func newQueryPage(jobID, pageToken string, prev *queryPage) *queryPage {
	return &queryPage{jobID, pageToken, prev}
}

func (p *queryPage) marshal() []byte {
	b, _ := json.Marshal(p)
	return b
}

func (p queryPage) encode() string {
	return base64.StdEncoding.EncodeToString(p.marshal())
}
