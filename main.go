package main

import (
	"github.com/Mwea/terraform-provider-chefsolo/chefsolo"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: chefsolo.Provider})
}
