package main

import (
	"flag"

	"github.com/enflow/terraform-provider-redirectpizza/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)
var (
	version string = "0.1.1"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug: debugMode,

		ProviderAddr: "github.com/enflow/redirectpizza",

		ProviderFunc: provider.New(version),
	}

	plugin.Serve(opts)
}
