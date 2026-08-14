package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = (*fileDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*fileDataSource)(nil)
)

type fileDataSource struct{ configuredDataSource }

func NewFileDataSource() datasource.DataSource { return &fileDataSource{} }

func (d *fileDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_file"
}

func (d *fileDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.configure(request, response)
}

func (d *fileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	attrs := fileDataAttributes()
	attrs["id"] = schema.StringAttribute{Required: true, MarkdownDescription: "Inode ID."}
	response.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Sigma file (inode) by ID. This is a read-only lookup; folders are managed with `sigma_folder`, and workbooks/reports remain data-source-only.",
		Attributes:          attrs,
	}
}

func (d *fileDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state fileDataModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	response.Diagnostics.Append(idDiags...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetFile(ctx, id)
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma file", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, fileData(value))...)
}
