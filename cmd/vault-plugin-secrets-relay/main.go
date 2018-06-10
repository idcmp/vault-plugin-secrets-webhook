package main

import (
	"os"

	"github.com/hashicorp/vault/helper/pluginutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/plugin"
	"github.com/idcmp/vault-plugin-secrets-relay"
	"log"
)

func main() {
	apiClientMeta := &pluginutil.APIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:]) // Ignore command, strictly parse flags

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := pluginutil.VaultPluginTLSProvider(tlsConfig)

	factoryFunc := relay.FactoryType(logical.TypeLogical)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: factoryFunc,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
