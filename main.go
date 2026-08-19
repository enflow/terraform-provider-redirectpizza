package main

import (
	"flag"

	"github.com/enflow/terraform-provider-redirectpizza/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

var (
	version string = "0.1.0"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.20.0 generate --provider-name=redirectpizza

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug: debugMode,

		ProviderAddr: "registry.terraform.io/enflow/redirectpizza",

		ProviderFunc: provider.New(version),
	}

	plugin.Serve(opts)
}
