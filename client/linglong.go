package client

import (
	"encoding/json"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

type linglongBuild struct {
	Branch        string `json:"branch"`
	CommentAuthor string `json:"comment_author"`
	GroupName     string `json:"group_name"`
	Project       string `json:"project"`
	RequestId     int    `json:"request_id"`
	AuthorEmail   string `json:"author_email"`
}

// triggerLinglong to build linglong
func (cl *Client) PostLinglongBuildjob() {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)

	author, email, _, err := cl.GetPRAuthorAndRef(GetOwner(), GetProject(), GetReqId())
	if err != nil {
		// Ignore failure
		log.Println("get pr author fail: ", err)
	}

	resp, err := client.R().
		SetBody(linglongBuild{
			Branch:        GetBranch(),
			GroupName:     GetOwner(),
			Project:       GetProject(),
			RequestId:     GetReqId(),
			CommentAuthor: author,
			AuthorEmail:   email,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/linglong/build")

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}
	if err != nil {
		log.Fatal(err)
	}

	var jobBuild JobTriggerJenkins
	err = json.Unmarshal([]byte(resp.Body()), &jobBuild)
	if err != nil {
		log.Fatal("trigger failed, response can't deserialize to jobBuild")
	}
	cl.id = jobBuild.ID
}
