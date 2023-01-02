package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"scaffolding_data_source": dataSourceScaffolding(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"redirectpizza_redirect": resourceRedirect(),
			},
			Schema: map[string]*schema.Schema{
				"token":        getAuthTokenSchema(),
				"api_base_url": getApiBaseUrlSchema(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	userAgent string
	baseUrl   string
	authToken string
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
		token := data.Get("token").(string)
		baseUrl := data.Get("api_base_url").(string)
		userAgent := p.UserAgent("terraform-provider-redirectpizza", version)

		return &apiClient{
			userAgent: userAgent,
			baseUrl:   baseUrl,
			authToken: token, // TODO: Validate that the token is set?
		}, nil
	}
}
