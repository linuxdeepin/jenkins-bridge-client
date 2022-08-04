package cmd

import (
	"context"
	"jenkins-bridge-client/client"
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	apipassCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	apipassCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	rootCmd.AddCommand(apipassCmd)
}

var apipassCmd = &cobra.Command{
	Use:   "apipass",
	Short: "apipass",
	Long:  `get api check status`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)

		owner := client.GetOwner()
		project := client.GetProject()

		pr, _, err := cl.GetGitHub().PullRequests.Get(context.Background(), owner, project, client.GetReqId())
		if err != nil {
			log.Fatal(err)
		}
		var body struct {
			Pass bool   `json:"pass"`
			Msg  string `json:"msg"`
		}
		params := map[string]string{
			"group":         owner,
			"project":       project,
			"branch":        client.GetBranch(),
			"request_id":    strconv.Itoa(client.GetReqId()),
			"request_event": client.GetEvent(),
			"commit_id":     pr.Head.GetSHA(),
		}
		log.Println("parmas", params)
		resp, err := cl.R().
			SetQueryParams(params).
			SetResult(&body).
			Get(cl.GetHost() + "/api/apicheck/status")
		if err != nil {
			log.Println(err)
		}
		if resp.StatusCode() != 200 {
			log.Fatal(resp.Status())
		}
		if !body.Pass {
			log.Fatal(body.Msg)
		}
	},
}
