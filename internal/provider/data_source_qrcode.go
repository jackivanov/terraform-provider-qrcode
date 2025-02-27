package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/skip2/go-qrcode"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &QRCodeDataSource{}

// QRCodeDataSource defines the QR code data source implementation.
type QRCodeDataSource struct{}

// NewQRCodeDataSource returns a new instance of QRCodeDataSource.
func NewQRCodeDataSource() datasource.DataSource {
	return &QRCodeDataSource{}
}

// Metadata returns the data source type name.
func (d *QRCodeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_generate"
}

// Schema defines the input and output attributes for the QR code data source.
func (d *QRCodeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"text": schema.StringAttribute{
				Description: "The text to encode as a QR code.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
							path.MatchRoot("text"),
							path.MatchRoot("sensitive_text"),
					),
			},
			},
			"sensitive_text": schema.StringAttribute{
				Description: "Sensitive text to encode as a QR code.",
				Sensitive:   true,
				Optional:    true,
			},
			"error_correction": schema.StringAttribute{
				Description: "Error correction level: L (low), M (medium, default), Q (high), H (highest).",
				Optional:    true,
			},
			"disable_border": schema.BoolAttribute{
				Description: "Set to true to disable the QR Code border.",
				Optional:    true,
			},
			"invert": schema.BoolAttribute{
				Description: "Set to true to invert black and white colors.",
				Optional:    true,
			},
			"ascii": schema.StringAttribute{
				Description: "ASCII text representation of the QR code.",
				Computed:    true,
			},
		},
	}
}

// Read generates the QR code in both Base64 PNG and ASCII formats.
func (d *QRCodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Define the input struct matching the schema
	var data struct {
		Text           types.String `tfsdk:"text"`
		SensitiveText  types.String `tfsdk:"sensitive_text"`
		ErrorCorrection types.String `tfsdk:"error_correction"`
		DisableBorder  types.Bool   `tfsdk:"disable_border"`
		Invert         types.Bool   `tfsdk:"invert"`
		ASCII          types.String `tfsdk:"ascii"`
	}

	// Read input data from Terraform
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine error correction level
	var level qrcode.RecoveryLevel
	switch strings.ToUpper(data.ErrorCorrection.ValueString()) {
	case "L":
		level = qrcode.Low
	case "M", "": // Default to Medium
		level = qrcode.Medium
	case "Q":
		level = qrcode.High
	case "H":
		level = qrcode.Highest
	default:
		resp.Diagnostics.AddError(
			"Invalid Error Correction Level",
			"Supported values: L (low), M (medium), Q (high), H (highest).",
		)
		return
	}

	// Determine which text to use for QR generation
	textValue := data.Text.ValueString()
	if data.Text.IsNull() {
		textValue = data.SensitiveText.ValueString()
	}

	// Generate QR code
	qr, err := qrcode.New(textValue, level)
	if err != nil {
		resp.Diagnostics.AddError(
			"QR Code Generation Failed",
			"Could not generate QR code: "+err.Error(),
		)
		return
	}

	// Apply optional flags
	if data.DisableBorder.ValueBool() {
		qr.DisableBorder = true
	}

	// Convert to ASCII (invert mode supported by the library)
	asciiQR := qr.ToSmallString(data.Invert.ValueBool()) // true = inverted mode

	// Set Terraform state
	data.ASCII = types.StringValue(asciiQR)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
