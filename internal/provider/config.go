package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func getAuthTokenSchema() *schema.Schema {
	return &schema.Schema{
		Description: "API token from the redirect.pizza dashboard (More → API).",
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
	}
}

func getApiBaseUrlSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Base URL for the redirect.pizza API.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "https://redirect.pizza/api/",
	}
}
