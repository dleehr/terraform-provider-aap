package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ action.Action = (*eventDispatch)(nil)
)

func NewEventDispatch() action.Action {
	return &eventDispatch{}
}

type eventDispatch struct{}

// metadata
func (a *eventDispatch) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eventdispatch"
}

// Schema
func (a *eventDispatch) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.UnlinkedSchema{
		Description: "Dispatches an event to AAP",
		Attributes: map[string]schema.Attribute{
			"limit": schema.StringAttribute{
				Description: "Ansible limit for job execution",
				Required:    true,
			},
			"template_type": schema.StringAttribute{
				Description: "Template type: job or workflow_job",
				Required:    true,
			},
			"job_template_name": schema.StringAttribute{
				Description: "Job Template Name",
				Optional:    true,
			},
			"workflow_job_template_name": schema.StringAttribute{
				Description: "Workflow Job Template Name",
				Optional:    true,
			},
			"organization_name": schema.StringAttribute{
				Description: "Organization Name",
				Required:    true,
			},
			"event_stream_config": schema.SingleNestedAttribute{
				Description: "Event Stream Configuration",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL",
						Required:    true,
						// Do we have sensitive for these attributes?
					},
					"username": schema.StringAttribute{
						Description: "Username",
						Required:    true,
					},
					"password": schema.StringAttribute{
						Description: "Password",
						Required:    true,
					},
				},
			},
		},
	}
}

type eventStreamConfig struct {
	Url      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type eventDispatchModel struct {
	Limit                   types.String      `tfsdk:"limit"`
	TemplateType            types.String      `tfsdk:"template_type"`
	JobTemplateName         types.String      `tfsdk:"job_template_name"`
	WorkflowJobTemplateName types.String      `tfsdk:"workflow_job_template_name"`
	OrganizationName        types.String      `tfsdk:"organization_name"`
	EventStreamConfig       eventStreamConfig `tfsdk:"event_stream_config"`
}

type payload struct {
	Limit                   string `json:"limit"`
	TemplateType            string `json:"template_type"`
	JobTemplateName         string `json:"job_template_name"`
	WorkflowJobTemplateName string `json:"workflow_job_template_name"`
	OrganizationName        string `json:"organization_name"`
}

func (a *eventDispatch) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	// Invoke the action
	var config eventDispatchModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sendPayload := &payload{
		TemplateType:            config.TemplateType.ValueString(),
		JobTemplateName:         config.JobTemplateName.ValueString(),
		WorkflowJobTemplateName: config.WorkflowJobTemplateName.ValueString(),
		OrganizationName:        config.OrganizationName.ValueString(),
		Limit:                   config.Limit.ValueString(),
	}

	jsonPayload, _ := json.Marshal(sendPayload)
	reader := bytes.NewReader(jsonPayload)

	url := config.EventStreamConfig.Url.ValueString()
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("\n About to POST event to %s \n\n", url),
	})

	contentType := "application/json"

	// Create the request
	hreq, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error creating request", err.Error()))
		return
	}

	hreq.Header.Set("Content-Type", contentType)
	hreq.SetBasicAuth(config.EventStreamConfig.Username.ValueString(), config.EventStreamConfig.Password.ValueString())
	client := &http.Client{}

	hresp, err := client.Do(hreq)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error sending request", err.Error()))
		return
	}
	defer hresp.Body.Close() // Close the response body when done

	body, err := ioutil.ReadAll(hresp.Body)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error reading response body", err.Error()))
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("\n Sent to POST event to %s, response status %s, response body %s\n\n", url, hresp.Status, body),
	})
}
