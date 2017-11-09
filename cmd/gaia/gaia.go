package main

import (
	"os"

	"github.com/spf13/cobra"
)

// GaiaCmd is the entry point for this binary
var GaiaCmd = &cobra.Command{
	Use:   "gaia",
	Short: "The Cosmos Network delegation-game blockchain test",
	Long:  "", //TODO re-work
	PersistenPreRunE: func(cmd *cobra.Command, args []string) error {
	},
}

func Execute() {
	addGlobalFlags()
	addGlobalCommands()
	GaiaCmd.Execute()
}

func addGlobalCommands() {

	//nodeCommands()
	//GaiaCmd.AddCommand(nodeCmd)

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
