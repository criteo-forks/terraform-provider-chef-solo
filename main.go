package main

import (
	"github.com/criteo-forks/terraform-provider-chefsolo/chefsolo"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: chefsolo.Provider})
}
