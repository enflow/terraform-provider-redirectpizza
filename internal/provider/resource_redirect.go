package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRedirect() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Sample resource in the Terraform provider scaffolding.",

		CreateContext: resourceRedirectCreate,
		ReadContext:   resourceRedirectRead,
		UpdateContext: resourceResourceUpdate,
		DeleteContext: resourceRedirectDelete,

		Schema: map[string]*schema.Schema{
			"sources": {
				Description: "The source domains (that the user enters in their browser).",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				MinItems: 1,
				MaxItems: 10, // TODO @ Mbardelmeijer
			},
			"destination": {
				Description: "The URL(s)where the user is redirected to.",
				Type:        schema.TypeList, // TODO Is this list ordered (TypeList) or unordered (TypeSet)? @ Mbardelmeijer
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Description: "the URL to redirect the user to",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
				Required: true,
				MinItems: 1,
				MaxItems: 1,
			},
			"redirect_type": { // TODO: Validate allowed values
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: redirectTypeValidator,
			},
		},
	}
}

func redirectTypeValidator(i interface{}, _ cty.Path) diag.Diagnostics {
	input, _ := i.(string)
	validRedirectTypes := []string{"permanent", "temporary", "frame", "permanent:308", "temporary:307"}

	for _, validRedirectType := range validRedirectTypes {
		if validRedirectType == input {
			return diag.Diagnostics{}
		}
	}
	return diag.Errorf("Invalid redirect type. Supported are: " + strings.Join(validRedirectTypes, ", "))
}

// https://redirect.pizza/docs#tag/Redirects/operation/createRedirect
func resourceRedirectCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	type httpCreateData struct {
		Sources         []string `json:"sources"`
		Destination     string   `json:"destination"`
		RedirectType    string   `json:"redirect_type"`
		UriForwarding   bool     `json:"uri_forwarding"`
		KeepQueryString bool     `json:"keep_query_string"`
		Tracking        bool     `json:"tracking"`
		Tags            []string `json:"tags"`
	}
	data := &httpCreateData{
		Sources:      []string{},
		RedirectType: d.Get("redirect_type").(string),

		// TODO: Do these defaults make sense?
		UriForwarding:   true,
		KeepQueryString: true,
		Tracking:        true,
		Tags:            []string{},
	}

	for _, source := range d.Get("sources").(*schema.Set).List() {
		data.Sources = append(data.Sources, source.(string))
	}

	for _, destination := range d.Get("destination").([]interface{}) {
		data.Destination = destination.(map[string]interface{})["url"].(string)
	}

	reqBody, _ := json.Marshal(data)

	apiClientData := meta.(*apiClient)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://redirect.pizza/api/v1/redirects", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+apiClientData.authToken)
	req.Header.Set("User-Agent", apiClientData.userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}

	if resp.StatusCode != 201 {
		return diag.Errorf("Expected status code 201 but got %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.Errorf("Cannot read response body: %s", err.Error())
	}

	respObj := &httpResponseData{}
	if err := json.Unmarshal(respBody, respObj); err != nil {
		return diag.Errorf("Cannot unmarshal response json: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%d", respObj.Data.Id))
	tflog.Trace(ctx, "Successfully created resource with id "+fmt.Sprintf("%d", respObj.Data.Id))
	return diag.Diagnostics{}
}

// TODO @Mbardelmeijer: De docs kloppen niet? De response zit in een 'data {}' blok die niet in de docs staan?
type httpResponseData struct {
	Data struct {
		Id      uint64 `json:"id"` // TODO @Mbardelmeijer: volgens de docs is dit een string, maar ik krijg een integer terug?
		Sources []struct {
			Id  uint64 `json:"id"` // TODO @Mbardelmeijer: volgens de docs is dit een string, maar ik krijg een integer terug?
			Url string `json:"url"`
		} `json:"sources"`
		Domains []struct {
			Id           int64  `json:"id"` // TODO @Mbardelmeijer: volgens de docs is dit een string, maar ik krijg een integer terug?
			Fqdn         string `json:"fqdn"`
			IsRootDomain bool   `json:"is_root_domain"`
			Dns          struct {
				Verified         bool `json:"verified"`
				RequiredSettings []struct { // TODO @mbardelmeijer: Volgens de docs is dit geen array?
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"required_settings"`
			} `json:"dns"`
			Security struct {
				Hsts                    bool `json:"hsts"`
				PreventForeignEmbedding bool `json:"prevent_foreign_embedding"`
			} `json:"security"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"domains"`
		Destination     string   `json:"destination"`
		RedirectType    string   `json:"redirect_type"`
		UriForwarding   bool     `json:"uri_forwarding"`
		KeepQueryString bool     `json:"keep_query_string"`
		Tracking        bool     `json:"tracking"`
		Tags            []string `json:"tags"`
	} `json:"data"`
}

// https://redirect.pizza/docs#tag/Redirects/operation/getRedirect
func resourceRedirectRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	apiClientData := meta.(*apiClient)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://redirect.pizza/api/v1/redirects/"+d.Id(), bytes.NewReader([]byte{}))
	req.Header.Set("Authorization", "Bearer "+apiClientData.authToken)
	req.Header.Set("User-Agent", apiClientData.userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return diag.Errorf("Expected http status 200, received: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.Errorf("could not read http response body: %s", err.Error())
	}

	respData := &httpResponseData{}
	if err := json.Unmarshal(body, respData); err != nil {
		return diag.Errorf("cannot read json from API: %s", err.Error())
	}

	d.SetId(d.Id())
	d.Set("destination.0.url", respData.Data.Destination)
	for i, src := range respData.Data.Sources {
		// TODO: Do we need/want to do antyhing with sources.%d.id?
		d.Set(fmt.Sprintf("sources.%d.url", i), src)
	}
	d.Set("redirect_type", respData.Data.RedirectType)

	// TODO: Other values like keep_query_string, tracking & tags
	return diag.Diagnostics{}
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("updating not implemented")
}

func resourceRedirectDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("delete not implemented")
}
