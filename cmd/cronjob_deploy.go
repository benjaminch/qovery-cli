package cmd

import (
	"context"
	"fmt"
	"github.com/qovery/qovery-cli/utils"
	"github.com/qovery/qovery-client-go"
	"github.com/spf13/cobra"
	"os"
)

var cronjobDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a cronjob",
	Run: func(cmd *cobra.Command, args []string) {
		utils.Capture(cmd)

		tokenType, token, err := utils.GetAccessToken()
		if err != nil {
			utils.PrintlnError(err)
			os.Exit(1)
		}

		client := utils.GetQoveryClient(tokenType, token)
		_, _, envId, err := getContextResourcesId(client)

		if err != nil {
			utils.PrintlnError(err)
			os.Exit(1)
		}

		if !utils.IsEnvironmentInATerminalState(envId, client) {
			utils.PrintlnError(fmt.Errorf("environment id '%s' is not in a terminal state. The request is not queued and you must wait "+
				"for the end of the current operation to run your command. Try again in a few moment", envId))
			os.Exit(1)
		}

		cronjobs, err := ListCronjobs(envId, client)

		if err != nil {
			utils.PrintlnError(err)
			os.Exit(1)
		}

		cronjob := utils.FindByJobName(cronjobs, cronjobName)

		if cronjob == nil {
			utils.PrintlnError(fmt.Errorf("cronjob %s not found", cronjobName))
			utils.PrintlnInfo("You can list all cronjobs with: qovery cronjob list")
			os.Exit(1)
		}

		docker := cronjob.Source.Docker.Get()
		image := cronjob.Source.Image.Get()

		var req qovery.JobDeployRequest

		if docker != nil {
			req = qovery.JobDeployRequest{
				GitCommitId: docker.GitRepository.DeployedCommitId,
			}

			if cronjobCommitId != "" {
				req.GitCommitId = &cronjobCommitId
			}
		} else {
			req = qovery.JobDeployRequest{
				ImageTag: image.Tag,
			}
		}

		_, _, err = client.JobActionsApi.DeployJob(context.Background(), cronjob.Id).JobDeployRequest(req).Execute()

		if err != nil {
			utils.PrintlnError(err)
			os.Exit(1)
		}

		utils.Println("Cronjob is deploying!")

		if watchFlag {
			utils.WatchJob(cronjob.Id, envId, client)
		}
	},
}

func init() {
	cronjobCmd.AddCommand(cronjobDeployCmd)
	cronjobDeployCmd.Flags().StringVarP(&organizationName, "organization", "", "", "Organization Name")
	cronjobDeployCmd.Flags().StringVarP(&projectName, "project", "", "", "Project Name")
	cronjobDeployCmd.Flags().StringVarP(&environmentName, "environment", "", "", "Environment Name")
	cronjobDeployCmd.Flags().StringVarP(&cronjobName, "cronjob", "n", "", "Cronjob Name")
	cronjobDeployCmd.Flags().StringVarP(&cronjobCommitId, "commit-id", "c", "", "Lifecycle Commit ID")
	cronjobDeployCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "Watch cronjob status until it's ready or an error occurs")

	_ = cronjobDeployCmd.MarkFlagRequired("cronjob")
}
