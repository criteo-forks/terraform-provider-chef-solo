package main

import (
	"github.com/criteo/terraform-provider-chef-solo/chef-solo-data"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: chef_solo_data.Provider})
}
