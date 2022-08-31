package cmd

import (
	"fmt"
	"jenkins-bridge-client/client"
	"strings"

	"github.com/spf13/cobra"
)

var arch string

func init() {
	isoBuildCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	isoBuildCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	isoBuildCmd.Flags().StringVarP(&arch, "arch", "", "amd", "iso build arch")
	rootCmd.AddCommand(isoBuildCmd)
}

var isoBuildCmd = &cobra.Command{
	Use:   "triggerISOBuild",
	Short: "trigger iso build",
	Long:  `trigger jenkins to build iso`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		tag := cl.GetLatestTagName(client.GetOwner(), client.GetProject())
		// get tag main version such as 20,23 and so on.
		tag = strings.Split(tag, ".")[0]
		cl.PostISOBuildJob(tag, arch)
		fmt.Println(cl.GetID())
	},
}
