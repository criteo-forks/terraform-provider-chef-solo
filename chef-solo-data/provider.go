package chef_solo_data

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"template_chef_solo": dataSourceChefSoloFile(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"template_chef_solo": schema.DataSourceResourceShim(
				"template_chef_solo",
				dataSourceChefSoloFile(),
			),
		},
	}
}
