package main

import (
	"github.com/spf13/cobra"
)

// GaiaCmd is the entry point for this binary
var GaiaCmd = &cobra.Command{
	Use:   "gaia",
	Short: "The Cosmos Network delegation-game blockchain test",
	Long:  "", //TODO re-work
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	addGlobalFlags()
	addGlobalCommands()
	GaiaCmd.Execute()
}

func addGlobalCommands() {

	nodeCommands()
	GaiaCmd.AddCommand(nodeCmd)

	serverCommands()
	GaiaCmd.AddCommand(serverCmd)

	//GaiaCmd.AddCommand()
	//GaiaCmd.AddCommand()
	//GaiaCmd.AddCommand()
	//GaiaCmd.AddCommand()
	//GaiaCmd.AddCommand()
}

func addGlobalFlags() {
	//GaiaCmd.PersistentFlags().StringVarP()
}
