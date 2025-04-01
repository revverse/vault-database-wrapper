package main

import (
	"log"
	"os"

	"vault-database-wrapper/plugin"
	"github.com/hashicorp/vault/api"
	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func main() {
	log.Println("Starting Vault Database Wrapper Plugin")
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	err := Run()
	log.Println("Plugin run completed")
	if err != nil {
		log.Println(err)
		log.Println("Exiting with error")
		os.Exit(1)
	}
}

func Run() error {
	dbplugin.ServeMultiplex(plugin.New)

	return nil
}