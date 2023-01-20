package main

import (
	"github.com/enflow/terraform-provider-redirectpizza/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)
var (
	version string = "0.1.1"
)

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	opts := &plugin.ServeOpts{
		ProviderAddr: "github.com/enflow/redirectpizza",

		ProviderFunc: provider.New(version),
	}

	plugin.Serve(opts)
}