package cmd

import (
	"fmt"
	"jenkins-bridge-client/client"

	"github.com/spf13/cobra"
)

var triggerSyncCmd = &cobra.Command{
	Use:   "triggerSync",
	Short: "trigger Archlinux build",
	Long:  `trigger jenkins to run build for Archlinux`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.PostApiJobSync()
		fmt.Println(cl.GetID())
	},
}

var apiCheckCmd = &cobra.Command{
	Use:   "triggerAbicheck",
	Short: "trigger api check",
	Long:  `trigger jenkins to run abi check`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.PostApiJobAbicheck()
		fmt.Println(cl.GetID())
	},
}

var triggerBuildCmd = &cobra.Command{
	Use:   "triggerBuild",
	Short: "trigger linglong build",
	Long:  `trigger jenkins to build deb`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.PostApiJobBuild()
		fmt.Println(cl.GetID())
	},
}

var archlinuxBuildCmd = &cobra.Command{
	Use:   "triggerArchlinux",
	Short: "trigger Archlinux build",
	Long:  `trigger jenkins to run build for Archlinux`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.PostApiJobArchlinux()
		fmt.Println(cl.GetID())
	},
}

func init() {
	apiCheckCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	apiCheckCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")

	triggerSyncCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	triggerSyncCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")

	triggerBuildCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	triggerBuildCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")

	archlinuxBuildCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	archlinuxBuildCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")

	rootCmd.AddCommand(apiCheckCmd)
	rootCmd.AddCommand(triggerSyncCmd)
	rootCmd.AddCommand(triggerBuildCmd)
	rootCmd.AddCommand(archlinuxBuildCmd)
}