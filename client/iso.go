package client

import (
	"encoding/json"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

type isoBuildRequest struct {
	Codename string `json:"codename"`
	Arch     string `json:"arch"`
}

var versionToCodeName = map[string]string{
	"20": "apricot",
	// system release start with 23 and tagged 1.0
	"1":  "beige",
	"23": "beige",
}

// triggerISO to build ISO
func (cl *Client) PostISOBuildJob(version, arch string) {
	client := resty.New()
	client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)

	resp, err := client.R().
		SetBody(isoBuildRequest{
			Codename: versionToCodeName[version],
			Arch:     arch,
		}).
		SetHeader("Accept", "application/json").
		SetHeader("X-token", cl.token).
		Post(cl.host + "/api/job/iso")

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != 200 {
		log.Fatal("trigger build fail, StatusCode not 200")
	}

	var jobBuild JobTriggerJenkins
	err = json.Unmarshal([]byte(resp.Body()), &jobBuild)
	if err != nil {
		log.Fatal("trigger failed, response can't deserialize to jobBuild")
	}
	cl.id = jobBuild.ID
}
