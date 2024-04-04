package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		Description: "A redirect is a resource that may contain multiple sources to a single destination.",

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
				Description: "The source domains (that the user enters in their browser).",
				Type:        schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				MinItems: 1,
				MaxItems: 1000,
			},

			"destination": {
				Description: "The URL(s)where the user is redirected to.",
				Type:        schema.TypeList, // The order of the destinations is relevant. Therefore this is a TypeList instead of a Set
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Description: "the URL to redirect the user to",
							Type:        schema.TypeString,
							Required:    true,
						},
						"expression": {
							Description: "The expression to evaluate for redirecting to the specified URL. Mandatory if multiple destinations are specified.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"monitoring": {
							Description: "The monitoring status for this destination. (must be one of: 'inherit' - default, 'enabled', 'disabled')",
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
				Description:      "The type of redirect to use.",
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
				Description: "Used to categorize redirects. May be an array or a string of comma-separated tags.",
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
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiClientData.baseUrl+"v1/redirects", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+apiClientData.authToken)
	req.Header.Set("User-Agent", apiClientData.userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}

	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return diag.Errorf("Expected status code 201 but got %d: %s", resp.StatusCode, string(respBody))
	}

	if err != nil {
		return diag.Errorf("Cannot read response body: %s", err.Error())
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
	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiClientData.baseUrl+"v1/redirects/"+d.Id(), bytes.NewReader([]byte{}))
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

	respData, err := parseApiResponse(body)
	if err != nil {
		return diag.Errorf("cannot parse api response: %v", err)
	}

	d.SetId(d.Id())
	for i, dst := range respData.Data.Destinations {
		d.Set(fmt.Sprintf("destination.%d.url", i), dst.Url)
		d.Set(fmt.Sprintf("destination.%d.expression", i), dst.Expression)
		d.Set(fmt.Sprintf("destination.%d.monitoring", i), dst.Monitoring)
	}

	sources := make([]interface{}, len(respData.Data.Sources), len(respData.Data.Sources))
	for i, src := range respData.Data.Sources {
		source := map[string]interface{}{
			"url": src.Url,
		}
		sources[i] = source
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
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", apiClientData.baseUrl+"v1/redirects/"+d.Id(), bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+apiClientData.authToken)
	req.Header.Set("User-Agent", apiClientData.userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}

	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return diag.Errorf("Expected status code 200 but got %d: %s", resp.StatusCode, string(respBody))
	}

	if err != nil {
		return diag.Errorf("Cannot read response body: %s", err.Error())
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
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", apiClientData.baseUrl+"v1/redirects/"+d.Id(), bytes.NewReader([]byte{}))
	req.Header.Set("Authorization", "Bearer "+apiClientData.authToken)
	req.Header.Set("User-Agent", apiClientData.userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return diag.Errorf("Cannot execute http request: %s", err.Error())
	}
	if resp.StatusCode != 204 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("<cannot read>")
		}

		return diag.Errorf("Expected http status 204, received: %d. Error: %s", resp.StatusCode, string(body))
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
