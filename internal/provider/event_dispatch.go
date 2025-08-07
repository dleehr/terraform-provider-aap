package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
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
			"name": schema.StringAttribute{
				Description: "Name of the event to dispatch",
			},
		},
	}
}

func (a *eventDispatch) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	// Invoke the action
	var config eventDispatchModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	eventDispatchName := config.Name.ValueString()

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("\n\nDispatch the event: %q\n\n", eventDispatchName),
	})

}

type eventDispatchModel struct {
	Name types.String `tfsdk:"name"`
}

// Invoke

// model struct
