package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRedirect() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a redirect.pizza redirect. One redirect can map many source hostnames to one or more destinations.",

		CreateContext: resourceRedirectCreate,
		ReadContext:   resourceRedirectRead,
		UpdateContext: resourceResourceUpdate,
		DeleteContext: resourceRedirectDelete,
		CustomizeDiff: destinationTypeValidator,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"sources": {
				Description: "Source hostnames or URLs visitors enter in their browser.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				MinItems: 1,
				MaxItems: 1000,
			},

			"destination": {
				Description: "Destination URLs. Order is significant for dynamic destinations. When more than one destination is set, all but the fallback must include an expression.",
				Type:        schema.TypeList, // The order of the destinations is relevant. Therefore this is a TypeList instead of a Set
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Description: "The URL to redirect the visitor to.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"expression": {
							Description: "Expression that selects this destination. Required on all but one destination when multiple destinations are set.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"monitoring": {
							Description: "Broken-destination monitoring for this URL. One of `inherit` (default), `enabled`, or `disabled`.",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "inherit",
						},
					},
				},
				Required: true,
				MinItems: 1,
			},

			"redirect_type": {
				Description:      "Redirect type. One of `permanent` (301), `temporary` (302), `permanent:308`, `temporary:307`, or `frame`.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: redirectTypeValidator,
				Default:          "permanent",
			},

			"keep_query_string": {
				Description: "Whether the query string should be forwarded to the destination URL.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"uri_forwarding": {
				Description: "Whether the path should be forwarded to the destination.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},

			"tracking": {
				Description: "Whether analytical information should be collected.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},

			"tags": {
				Description: "Tags used to categorize this redirect.",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func destinationTypeValidator(ctx context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	dstDiff := diff.Get("destination").([]interface{})
	if len(dstDiff) < 2 {
		return nil
	}

	expressionCount := 0
	for _, v := range dstDiff {
		vv := v.(map[string]interface{})
		if expression, set := vv["expression"]; set && expression != "" {
			expressionCount++
		}
	}

	if expressionCount < len(dstDiff)-1 {
		return fmt.Errorf("not all destinations have an expression specified but multiple destinations were defined")
	}

	return nil
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
	data := hydrateHttpPersistData(d)
	merge := false // Do not use the merge functionality for resources managed by Terraform
	data.Merge = &merge
	reqBody, _ := json.Marshal(data)

	apiClientData := meta.(*apiClient)
	status, respBody, err := apiClientData.do(ctx, http.MethodPost, "v1/redirects", reqBody)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}
	if status != http.StatusCreated {
		return diag.Errorf("Expected status code 201 but got %d: %s", status, string(respBody))
	}

	respObj, err := parseApiResponse(respBody)
	if err != nil {
		return diag.Errorf("cannot parse api response: %v", err)
	}

	d.SetId(fmt.Sprintf("%d", respObj.Data.Id))
	tflog.Trace(ctx, "Successfully created resource with id "+fmt.Sprintf("%d", respObj.Data.Id))
	return diag.Diagnostics{}
}

type httpResponseData struct {
	Data struct {
		Id      uint64 `json:"id"`
		Sources []struct {
			Id  uint64 `json:"id"`
			Url string `json:"url"`
		} `json:"sources"`
		Domains []struct {
			Id           int64  `json:"id"`
			Fqdn         string `json:"fqdn"`
			IsRootDomain bool   `json:"is_root_domain"`
			Dns          struct {
				Verified         bool `json:"verified"`
				RequiredSettings []struct {
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
		DestinationJson *json.RawMessage `json:"destination"`
		Destinations    []httpDestination
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
	status, body, err := apiClientData.do(ctx, http.MethodGet, "v1/redirects/"+d.Id(), nil)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}
	if status != http.StatusOK {
		return diag.Errorf("Expected http status 200, received: %d", status)
	}

	respData, err := parseApiResponse(body)
	if err != nil {
		return diag.Errorf("cannot parse api response: %v", err)
	}

	d.SetId(d.Id())
	destinations := []map[string]string{}
	for _, dst := range respData.Data.Destinations {
		if dst.Monitoring == "" {
			dst.Monitoring = "inherit"
		}

		destinations = append(destinations, map[string]string{
			"url":        dst.Url,
			"expression": dst.Expression,
			"monitoring": dst.Monitoring,
		})
	}
	d.Set("destination", destinations)

	sources := []string{}
	for _, src := range respData.Data.Sources {
		sources = append(sources, src.Url)
	}
	d.Set("sources", sources)
	d.Set("redirect_type", respData.Data.RedirectType)
	d.Set("keep_query_string", respData.Data.KeepQueryString)
	d.Set("uri_forwarding", respData.Data.UriForwarding)
	d.Set("tracking", respData.Data.Tracking)
	d.Set("tags", respData.Data.Tags)

	return diag.Diagnostics{}
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	data := hydrateHttpPersistData(d)
	reqBody, _ := json.Marshal(data)

	apiClientData := meta.(*apiClient)
	status, respBody, err := apiClientData.do(ctx, http.MethodPut, "v1/redirects/"+d.Id(), reqBody)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}
	if status != http.StatusOK {
		return diag.Errorf("Expected status code 200 but got %d: %s", status, string(respBody))
	}

	respObj, err := parseApiResponse(respBody)
	if err != nil {
		return diag.Errorf("cannot parse api response: %v", err)
	}

	d.SetId(fmt.Sprintf("%d", respObj.Data.Id))
	tflog.Trace(ctx, "Successfully updated resource with id "+fmt.Sprintf("%d", respObj.Data.Id))
	return diag.Diagnostics{}
}

type httpDestination struct {
	Url        string `json:"url"`
	Expression string `json:"expression"`
	Monitoring string `json:"monitoring"`
}

type httpPersistData struct {
	Sources         []string          `json:"sources"`
	Destinations    []httpDestination `json:"destination"`
	RedirectType    string            `json:"redirect_type"`
	UriForwarding   bool              `json:"uri_forwarding"`
	KeepQueryString bool              `json:"keep_query_string"`
	Tracking        bool              `json:"tracking"`
	Tags            []string          `json:"tags"`
	Merge           *bool             `json:"merge,omitempty"`
}

func hydrateHttpPersistData(d *schema.ResourceData) *httpPersistData {
	tags := []string{}
	for _, tag := range d.Get("tags").(*schema.Set).List() {
		tags = append(tags, tag.(string))
	}
	data := &httpPersistData{
		Sources:      []string{},
		RedirectType: d.Get("redirect_type").(string),

		UriForwarding:   d.Get("uri_forwarding").(bool),
		KeepQueryString: d.Get("keep_query_string").(bool),
		Tracking:        d.Get("tracking").(bool),
		Tags:            tags,
	}

	for _, source := range d.Get("sources").(*schema.Set).List() {
		data.Sources = append(data.Sources, source.(string))
	}
	for _, destination := range d.Get("destination").([]interface{}) {
		data.Destinations = append(data.Destinations, httpDestination{
			Url:        destination.(map[string]interface{})["url"].(string),
			Expression: destination.(map[string]interface{})["expression"].(string),
			Monitoring: destination.(map[string]interface{})["monitoring"].(string),
		})
	}
	return data
}

func resourceRedirectDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	apiClientData := meta.(*apiClient)
	status, body, err := apiClientData.do(ctx, http.MethodDelete, "v1/redirects/"+d.Id(), nil)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}
	if status != http.StatusNoContent {
		return diag.Errorf("Expected http status 204, received: %d. Error: %s", status, string(body))
	}

	d.SetId("")
	return diag.Diagnostics{}
}

func parseApiResponse(respBody []byte) (*httpResponseData, error) {
	respObj := &httpResponseData{}
	if err := json.Unmarshal(respBody, respObj); err != nil {
		return nil, fmt.Errorf("Cannot unmarshal response json: %v", err)
	}

	if err := json.Unmarshal(*respObj.Data.DestinationJson, &respObj.Data.Destinations); err != nil {
		dst := httpDestination{}
		if err2 := json.Unmarshal(*respObj.Data.DestinationJson, &dst.Url); err2 != nil {
			return nil, fmt.Errorf("Cannot parse Destination as either a string (%v) or object (%v)", err2, err)
		}
		respObj.Data.Destinations = append(respObj.Data.Destinations, dst)
	}

	return respObj, nil
}
