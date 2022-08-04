package cmd

import (
	"jenkins-bridge-client/client"

	"github.com/spf13/cobra"
)

var printLogCmd = &cobra.Command{
	Use:   "printLog",
	Short: "get job output",
	Long:  `sync jenkins job's console output`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.SetID(runid)
		cl.PrintLog()
	},
}

var cancelBuildCmd = &cobra.Command{
	Use:   "cancelBuild",
	Short: "trigger job canceled",
	Long:  `make jenkins jobs canceled`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.SetID(runid)
		cl.GetApiJobCancel()
	},
}

var downloadArtifactsCmd = &cobra.Command{
	Use:   "downloadArtifacts",
	Short: "download jenkins result",
	Long:  `trigger jenkins to build linglong`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.SetID(runid)
		cl.DownloadArtifacts()
	},
}

func init() {
	printLogCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	printLogCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	printLogCmd.Flags().IntVarP(&runid, "runid", "", 0, "")
	printLogCmd.MarkFlagRequired("runid")

	cancelBuildCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	cancelBuildCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	cancelBuildCmd.Flags().IntVarP(&runid, "runid", "", 0, "")
	cancelBuildCmd.MarkFlagRequired("runid")

	downloadArtifactsCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	downloadArtifactsCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	downloadArtifactsCmd.MarkFlagRequired("runid")

	rootCmd.AddCommand(printLogCmd)
	rootCmd.AddCommand(cancelBuildCmd)
	rootCmd.AddCommand(downloadArtifactsCmd)
}
