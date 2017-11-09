package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/commands"
	rest "github.com/cosmos/cosmos-sdk/client/rest"
	coinrest "github.com/cosmos/cosmos-sdk/modules/coin/rest"
	noncerest "github.com/cosmos/cosmos-sdk/modules/nonce/rest"
	rolerest "github.com/cosmos/cosmos-sdk/modules/roles/rest"

	"github.com/tendermint/tmlibs/cli"

	stakerest "github.com/cosmos/gaia/modules/stake/rest"
)

// XXX [zr] how/where to configure...?

// this should share the dir with gaiacli, so you can use the cli and
// the api interchangeably
// cmd := cli.PrepareMainCmd(srvCli, "BC", os.ExpandEnv("$HOME/.gaiacli"))

var portFlag = "port"

const defaultAlgo = "ed25519"

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Light REST client for tendermint",
	Long:  `Gaiaserver presents  a nice (not raw hex) interface to the gaia blockchain structure.`,
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the light REST client for tendermint",
	Long:  "Access gaia via REST",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdServe(cmd, args)
	},
}

func serverCommands() {
	serverCmd.AddCommand(commands.InitCmd)
	serverCmd.AddCommand(commands.VersionCmd)
	serverCmd.AddCommand(serveCmd)
	addServerFlags()
}

func addServerFlags() {
	commands.AddBasicFlags(srvCli)
	serveCmd.PersistentFlags().Int(envPortFlag, 8998, "port to run the server on")
}

func cmdServe(cmd *cobra.Command, args []string) error {
	router := mux.NewRouter()

	routeRegistrars := []func(*mux.Router) error{
		// rest.Keys handlers
		rest.NewDefaultKeysManager(defaultAlgo).RegisterAllCRUD,

		// Coin send handler
		coinrest.RegisterCoinSend,
		// Coin query account handler
		coinrest.RegisterQueryAccount,

		// Roles createRole handler
		rolerest.RegisterCreateRole,

		// Gaia sign transactions handler
		rest.RegisterSignTx,
		// Gaia post transaction handler
		rest.RegisterPostTx,

		// Nonce query handler
		noncerest.RegisterQueryNonce,

		//staking query functionality TODO add tx and all query here
		stakerest.RegisterQueryCandidate,
	}

	for _, routeRegistrar := range routeRegistrars {
		if err := routeRegistrar(router); err != nil {
			log.Fatal(err)
		}
	}

	addr := fmt.Sprintf(":%d", portFlag)

	log.Printf("Serving on %q", addr)
	return http.ListenAndServe(addr, router)
}
