package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &EventStreamDataSource{}

func NewEventStreamDataSource() datasource.DataSource {
	return &EventStreamDataSource{
		client: nil,
	}
}

type EventStreamDataSource struct {
	client ProviderHTTPClient
}

type EventStreamDataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
}

func (d *EventStreamDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eventstream"
}

func (d *EventStreamDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"url": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *EventStreamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Check that the response and diagnostics pointer is defined
	if resp == nil {
		tflog.Error(ctx, "Response not defined, we cannot continue with the execution")
		return
	}

	// Check that the current context is active
	if !IsContextActive("Configure", ctx, &resp.Diagnostics) {
		return
	}

	// Check that the provider data is configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*AAPClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AAPClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// JSON for the data source
type EventStreamAPIModel struct {
	Name string `json:"name"`
	Id   int64  `json:"id"`
	URL  string `json:"url"`
}

type EventStreamAPIModelList struct {
	Results []EventStreamAPIModel `json:"results"`
}

func (d *EventStreamDataSourceModel) ParseHttpResponseList(body []byte) diag.Diagnostics {
	var diags diag.Diagnostics

	// Unmarshal the JSON response
	var apiModelList EventStreamAPIModelList
	err := json.Unmarshal(body, &apiModelList)
	if err != nil {
		diags.AddError("Error parsing JSON response from AAP", err.Error())
		return diags
	}

	if len(apiModelList.Results) != 1 {
		diags.AddError("Unable to fetch event_stream from AAP", fmt.Sprintf("Expected 1 object in JSON response, found %d", len(apiModelList.Results)))
		return diags
	}

	var apiModel = apiModelList.Results[0]

	d.ID = types.Int64Value(apiModel.Id)
	d.URL = ParseStringValue(apiModel.URL)
	d.Name = ParseStringValue(apiModel.Name)
	return diags
}

func (d *EventStreamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EventStreamDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	edaEndpoint := strings.Replace(d.client.getApiEndpoint(), "controller/v2", "eda/v1", 1)

	// resourceURL := path.Join(edaEndpoint, "event-streams", fmt.Sprint(state.ID.ValueInt64()))
	// queryURL := path.Join(edaEndpoint, fmt.Sprint("event-streams?name=", state.Name.String()))
	url := path.Join(edaEndpoint, "event-streams")
	// queryURL := fmt.Sprint(url, "?name=", state.Name.ValueString())

	// Look up by name and org, this will return a list and we assume the first is valid
	// readResponseBody, diags := d.client.Get(queryURL)
	var query = map[string]string{
		"name": state.Name.ValueString(),
	}
	response, body, err := d.client.doRequest(http.MethodGet, url, query, nil)
	diags := ValidateResponse(response, body, err, []int{http.StatusOK})

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = state.ParseHttpResponseList(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
