package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/skip2/go-qrcode"

	"path/filepath"
)

// Ensure implementation satisfies the expected interfaces
var _ resource.Resource = &qrcodeResource{}

// qrcodeResource is the resource implementation.
type qrcodeResource struct{}

// NewQRCodeResource creates a new QR code resource instance.
func NewQRCodeResource() resource.Resource {
	return &qrcodeResource{}
}

// Metadata returns the resource type name.
func (r *qrcodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_generate"
}

// Schema defines the resource schema.
func (r *qrcodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"text": schema.StringAttribute{
				Optional:    true,
				Description: "The text content to encode in the QR code.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("text"),
						path.MatchRoot("sensitive_text"),
					),
				},
			},
			"sensitive_text": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Sensitive text content to encode in the QR code.",
			},
			"size": schema.Int64Attribute{
				Optional:    true,
				Description: "Size of the QR code image in pixels.",
			},
			"file": schema.StringAttribute{
				Required:    true,
				Description: "Path to save the generated QR code image.",
			},
			"sha256": schema.StringAttribute{
				Computed:    true,
				Description: "SHA-256 checksum of the generated QR code image.",
			},
		},
	}
}

// Create generates a QR code and saves it to a file.
func (r *qrcodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan struct {
		Text          types.String `tfsdk:"text"`
		SensitiveText types.String `tfsdk:"sensitive_text"`
		Size          types.Int64  `tfsdk:"size"`
		File          types.String `tfsdk:"file"`
		SHA256        types.String `tfsdk:"sha256"`
	}

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine which text to use
	qrText := ""
	if !plan.Text.IsNull() {
		qrText = plan.Text.ValueString()
	} else {
		qrText = plan.SensitiveText.ValueString()
	}

	// Set size
	const defaultSize = 256
	const minSize = 100
	const maxSize = 2000

	size := defaultSize
	if !plan.Size.IsNull() {
		sizeVal := int(plan.Size.ValueInt64())
		if sizeVal < minSize || sizeVal > maxSize {
			resp.Diagnostics.AddError("Invalid Size", fmt.Sprintf("Size must be between %d and %d pixels.", minSize, maxSize))
			return
		}
		size = sizeVal
	}

	// Generate QR code
	pngData, err := qrcode.Encode(qrText, qrcode.Medium, size)
	if err != nil {
		resp.Diagnostics.AddError("QR Code Generation Failed", err.Error())
		return
	}

	// Compute SHA-256 checksum
	hash := sha256.Sum256(pngData)
	sha256Checksum := hex.EncodeToString(hash[:])

	// Save to file
	filePath := plan.File.ValueString()
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		resp.Diagnostics.AddError("Failed to Create Directory", err.Error())
		return
	}

	err = os.WriteFile(filePath, pngData, 0644)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Save QR Code", err.Error())
		return
	}

	// Set state
	resp.State.Set(ctx, &struct {
		Text          types.String `tfsdk:"text"`
		SensitiveText types.String `tfsdk:"sensitive_text"`
		Size          types.Int64  `tfsdk:"size"`
		File          types.String `tfsdk:"file"`
		SHA256        types.String `tfsdk:"sha256"`
	}{
		Text:          plan.Text,
		SensitiveText: plan.SensitiveText,
		Size:          plan.Size,
		File:          plan.File,
		SHA256:        types.StringValue(sha256Checksum),
	})
}

// Read refreshes the state.
func (r *qrcodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state struct {
		Text          types.String `tfsdk:"text"`
		SensitiveText types.String `tfsdk:"sensitive_text"`
		Size          types.Int64  `tfsdk:"size"`
		File          types.String `tfsdk:"file"`
		SHA256        types.String `tfsdk:"sha256"`
	}

	// Read the state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the file path is not set, nothing to check
	if state.File.IsNull() || state.File.ValueString() == "" {
		return
	}

	filePath := state.File.ValueString()

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File is missing, remove the resource from the state
		resp.State.RemoveResource(ctx)
	}
}

// Update is identical to Create since QR codes are immutable.
func (r *qrcodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.Create(ctx, resource.CreateRequest{
		Plan: req.Plan,
	}, (*resource.CreateResponse)(resp))
}

// Delete removes the QR code file and the resource from state.
func (r *qrcodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state struct {
			Text          types.String `tfsdk:"text"`
			SensitiveText types.String `tfsdk:"sensitive_text"`
			Size          types.Int64  `tfsdk:"size"`
			File          types.String `tfsdk:"file"`
			SHA256        types.String `tfsdk:"sha256"`
	}

	// Read current state
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
			return
	}

	// Remove the file if it exists
	if state.File.IsNull() || state.File.ValueString() == "" {
    return // No file to delete
	}

	filePath := state.File.ValueString()

	if _, err := os.Stat(filePath); err == nil {
			// File exists, attempt to delete
			if err := os.Remove(filePath); err != nil {
					resp.Diagnostics.AddError("Failed to Delete QR Code", err.Error())
					return
			}
	}

	// Remove the resource from state
	resp.State.RemoveResource(ctx)
}

