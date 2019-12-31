package chefsolo

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"chefsolo": dataSourceChefSoloFile(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"chefsolo": schema.DataSourceResourceShim(
				"chefsolo",
				dataSourceChefSoloFile(),
			),
		},
	}
}
