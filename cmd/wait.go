package cmd

import (
	"jenkins-bridge-client/client"
	"log"

	"github.com/spf13/cobra"
)

func init() {
	waitCmd.Flags().StringVarP(&token, "token", "", defaultToken, "jenkins bridge token")
	waitCmd.Flags().StringVarP(&server, "server", "", defaultServer, "jenkins bridge server address")
	waitCmd.Flags().IntVarP(&runid, "runid", "", 0, "jenkins bridge server address")
	waitCmd.MarkFlagRequired("runid")
	rootCmd.AddCommand(waitCmd)
}

var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "trigger linglong build",
	Long:  `trigger jenkins to build linglong`,
	Run: func(cmd *cobra.Command, args []string) {
		cl := client.NewClient()
		cl.SetHost(server)
		cl.SetToken(token)
		cl.SetID(runid)
		err := cl.Wait()
		if err != nil {
			log.Fatal(err)
		}
	},
}
