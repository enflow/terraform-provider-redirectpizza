package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func getAuthTokenSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Optional:    false,
		Computed:    false,
		ForceNew:    false,
		Sensitive:   true,
	}
}

func getApiBaseUrlSchema() *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeString,
		Optional:  true,
		Sensitive: false,
		Default:   "https://redirect.pizza/api/",
	}
}
