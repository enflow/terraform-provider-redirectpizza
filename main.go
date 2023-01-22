package main

import (
	"github.com/enflow/terraform-provider-redirectpizza/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

var (
	version string = "0.1.1"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	opts := &plugin.ServeOpts{
		ProviderAddr: "github.com/enflow/redirectpizza",

		ProviderFunc: provider.New(version),
	}

	plugin.Serve(opts)
}
