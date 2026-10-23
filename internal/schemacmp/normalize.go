package schemacmp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	pschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Raw TF schema JSON wire types matching OpenTofu / Terraform providers schema -json output.
type rawTFSchemaDump struct {
	FormatVersion   string                            `json:"format_version"`
	ProviderSchemas map[string]rawTFProviderSchemaDoc `json:"provider_schemas"`
}

type rawTFProviderSchemaDoc struct {
	Provider          *rawTFSchemaBlock           `json:"provider,omitempty"`
	ResourceSchemas   map[string]rawTFSchemaBlock `json:"resource_schemas,omitempty"`
	DataSourceSchemas map[string]rawTFSchemaBlock `json:"data_source_schemas,omitempty"`
}

type rawTFSchemaBlock struct {
	Version int64       `json:"version,omitempty"`
	Block   *rawTFBlock `json:"block,omitempty"`
}

type rawTFBlock struct {
	Attributes map[string]rawTFAttribute `json:"attributes,omitempty"`
	BlockTypes map[string]rawTFBlockType `json:"block_types,omitempty"`
	Deprecated bool                      `json:"deprecated,omitempty"`
}

type rawTFAttribute struct {
	Type       any              `json:"type,omitempty"`
	Required   bool             `json:"required,omitempty"`
	Optional   bool             `json:"optional,omitempty"`
	Computed   bool             `json:"computed,omitempty"`
	Sensitive  bool             `json:"sensitive,omitempty"`
	Deprecated bool             `json:"deprecated,omitempty"`
	NestedType *rawTFNestedType `json:"nested_type,omitempty"`
}

type rawTFNestedType struct {
	Attributes  map[string]rawTFAttribute `json:"attributes,omitempty"`
	NestingMode string                    `json:"nesting_mode,omitempty"`
	MinItems    int64                     `json:"min_items,omitempty"`
	MaxItems    int64                     `json:"max_items,omitempty"`
}

type rawTFBlockType struct {
	NestingMode string      `json:"nesting_mode,omitempty"`
	Block       *rawTFBlock `json:"block,omitempty"`
	MinItems    int64       `json:"min_items,omitempty"`
	MaxItems    int64       `json:"max_items,omitempty"`
}

// NormalizeTFJSON parses OpenTofu / Terraform providers schema -json output into a NormalizedSchema.
func NormalizeTFJSON(data []byte) (*NormalizedSchema, error) {
	var dump rawTFSchemaDump
	if err := json.Unmarshal(data, &dump); err != nil {
		return nil, fmt.Errorf("failed to parse terraform schema json: %w", err)
	}

	if len(dump.ProviderSchemas) == 0 {
		return nil, fmt.Errorf("terraform schema json contains no provider_schemas")
	}

	// Locate the seerr provider doc
	var seerrDoc *rawTFProviderSchemaDoc
	for addr, doc := range dump.ProviderSchemas {
		if strings.Contains(addr, "seerr") {
			docCopy := doc
			seerrDoc = &docCopy
			break
		}
	}
	if seerrDoc == nil {
		// Fallback to the first provider found
		for _, doc := range dump.ProviderSchemas {
			docCopy := doc
			seerrDoc = &docCopy
			break
		}
	}

	out := &NormalizedSchema{
		FormatVersion: "1.0",
		Resources:     make(map[string]*NormalizedBlock),
		DataSources:   make(map[string]*NormalizedBlock),
	}

	if seerrDoc.Provider != nil && seerrDoc.Provider.Block != nil {
		out.Provider = normalizeRawBlock(seerrDoc.Provider.Block)
	}

	for name, resDoc := range seerrDoc.ResourceSchemas {
		if resDoc.Block != nil {
			out.Resources[name] = normalizeRawBlock(resDoc.Block)
		}
	}

	for name, dsDoc := range seerrDoc.DataSourceSchemas {
		if dsDoc.Block != nil {
			out.DataSources[name] = normalizeRawBlock(dsDoc.Block)
		}
	}

	return out, nil
}

func normalizeRawBlock(b *rawTFBlock) *NormalizedBlock {
	if b == nil {
		return &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
	}

	nb := &NormalizedBlock{
		Attributes: make(map[string]*NormalizedAttribute),
		Blocks:     make(map[string]*NormalizedBlockType),
	}

	for name, attr := range b.Attributes {
		nb.Attributes[name] = normalizeRawAttribute(attr)
	}

	for name, bt := range b.BlockTypes {
		nb.Blocks[name] = &NormalizedBlockType{
			NestingMode: bt.NestingMode,
			MinItems:    bt.MinItems,
			MaxItems:    bt.MaxItems,
			Block:       normalizeRawBlock(bt.Block),
		}
	}

	return nb
}

func normalizeRawAttribute(attr rawTFAttribute) *NormalizedAttribute {
	na := &NormalizedAttribute{
		Type:       CanonicalizeType(attr.Type),
		Required:   attr.Required,
		Optional:   attr.Optional,
		Computed:   attr.Computed,
		Sensitive:  attr.Sensitive,
		Deprecated: attr.Deprecated,
	}

	if attr.NestedType != nil {
		na.NestedType = &NormalizedNestedType{
			NestingMode: attr.NestedType.NestingMode,
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range attr.NestedType.Attributes {
			na.NestedType.Attributes[name] = normalizeRawAttribute(child)
		}
	}

	return na
}

// CanonicalizeType produces a normalized string for primitive, collection, and structural types.
func CanonicalizeType(raw any) string {
	if raw == nil {
		return ""
	}

	switch v := raw.(type) {
	case string:
		return normalizePrimitiveTypeName(v)
	case []any:
		if len(v) == 0 {
			return "dynamic"
		}
		kind, ok := v[0].(string)
		if !ok {
			return fmt.Sprintf("%v", v)
		}
		switch kind {
		case "list", "set", "map":
			elemType := "dynamic"
			if len(v) > 1 {
				elemType = CanonicalizeType(v[1])
			}
			return fmt.Sprintf("%s(%s)", kind, elemType)
		case "object":
			if len(v) > 1 {
				if objMap, ok := v[1].(map[string]any); ok {
					var keys []string
					for k := range objMap {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					var parts []string
					for _, k := range keys {
						parts = append(parts, fmt.Sprintf("%s:%s", k, CanonicalizeType(objMap[k])))
					}
					return fmt.Sprintf("object({%s})", strings.Join(parts, ","))
				}
			}
			return "object({})"
		case "tuple":
			if len(v) > 1 {
				if tupleList, ok := v[1].([]any); ok {
					var parts []string
					for _, item := range tupleList {
						parts = append(parts, CanonicalizeType(item))
					}
					return fmt.Sprintf("tuple([%s])", strings.Join(parts, ","))
				}
			}
			return "tuple([])"
		default:
			return fmt.Sprintf("%v", v)
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}

func normalizePrimitiveTypeName(t string) string {
	switch t {
	case "int64", "float64", "number":
		return "number"
	default:
		return t
	}
}

// NormalizeProvider extracts and normalizes the schema directly from a Go provider.Provider.
func NormalizeProvider(p provider.Provider) (*NormalizedSchema, error) {
	ctx := context.Background()

	var pResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &pResp)
	if pResp.Diagnostics.HasError() {
		return nil, fmt.Errorf("provider Schema() diagnostics: %v", pResp.Diagnostics)
	}

	out := &NormalizedSchema{
		FormatVersion: "1.0",
		Resources:     make(map[string]*NormalizedBlock),
		DataSources:   make(map[string]*NormalizedBlock),
	}

	out.Provider = normalizeProviderSchema(pResp.Schema)

	// Resources
	for _, factory := range p.Resources(ctx) {
		r := factory()
		var mResp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "seerr"}, &mResp)

		var sResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &sResp)
		if sResp.Diagnostics.HasError() {
			return nil, fmt.Errorf("resource %q Schema() diagnostics: %v", mResp.TypeName, sResp.Diagnostics)
		}

		out.Resources[mResp.TypeName] = normalizeResourceSchema(sResp.Schema)
	}

	// Data Sources
	for _, factory := range p.DataSources(ctx) {
		ds := factory()
		var mResp datasource.MetadataResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "seerr"}, &mResp)

		var sResp datasource.SchemaResponse
		ds.Schema(ctx, datasource.SchemaRequest{}, &sResp)
		if sResp.Diagnostics.HasError() {
			return nil, fmt.Errorf("datasource %q Schema() diagnostics: %v", mResp.TypeName, sResp.Diagnostics)
		}

		out.DataSources[mResp.TypeName] = normalizeDataSourceSchema(sResp.Schema)
	}

	return out, nil
}

func normalizeProviderSchema(s pschema.Schema) *NormalizedBlock {
	nb := &NormalizedBlock{
		Attributes: make(map[string]*NormalizedAttribute),
		Blocks:     make(map[string]*NormalizedBlockType),
	}
	for name, attr := range s.Attributes {
		nb.Attributes[name] = normalizeFrameworkProviderAttribute(attr)
	}
	for name, blk := range s.Blocks {
		nb.Blocks[name] = normalizeFrameworkProviderBlock(blk)
	}
	return nb
}

func normalizeResourceSchema(s rschema.Schema) *NormalizedBlock {
	nb := &NormalizedBlock{
		Attributes: make(map[string]*NormalizedAttribute),
		Blocks:     make(map[string]*NormalizedBlockType),
	}
	for name, attr := range s.Attributes {
		nb.Attributes[name] = normalizeFrameworkResourceAttribute(attr)
	}
	for name, blk := range s.Blocks {
		nb.Blocks[name] = normalizeFrameworkResourceBlock(blk)
	}
	return nb
}

func normalizeDataSourceSchema(s dschema.Schema) *NormalizedBlock {
	nb := &NormalizedBlock{
		Attributes: make(map[string]*NormalizedAttribute),
		Blocks:     make(map[string]*NormalizedBlockType),
	}
	for name, attr := range s.Attributes {
		nb.Attributes[name] = normalizeFrameworkDataSourceAttribute(attr)
	}
	for name, blk := range s.Blocks {
		nb.Blocks[name] = normalizeFrameworkDataSourceBlock(blk)
	}
	return nb
}

func normalizeFrameworkResourceAttribute(a rschema.Attribute) *NormalizedAttribute {
	na := &NormalizedAttribute{
		Required:   a.IsRequired(),
		Optional:   a.IsOptional(),
		Computed:   a.IsComputed(),
		Sensitive:  a.IsSensitive(),
		Deprecated: a.GetDeprecationMessage() != "",
	}

	switch typed := a.(type) {
	case rschema.StringAttribute:
		na.Type = "string"
	case rschema.BoolAttribute:
		na.Type = "bool"
	case rschema.Int64Attribute, rschema.Float64Attribute, rschema.NumberAttribute:
		na.Type = "number"
	case rschema.ListAttribute:
		na.Type = fmt.Sprintf("list(%s)", frameworkTypeToString(typed.ElementType))
	case rschema.SetAttribute:
		na.Type = fmt.Sprintf("set(%s)", frameworkTypeToString(typed.ElementType))
	case rschema.MapAttribute:
		na.Type = fmt.Sprintf("map(%s)", frameworkTypeToString(typed.ElementType))
	case rschema.ListNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "list",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.NestedObject.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkResourceAttribute(child)
		}
	case rschema.SetNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "set",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.NestedObject.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkResourceAttribute(child)
		}
	case rschema.MapNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "map",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.NestedObject.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkResourceAttribute(child)
		}
	case rschema.SingleNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "single",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkResourceAttribute(child)
		}
	default:
		na.Type = CanonicalizeType(typed.GetType().String())
	}

	return na
}

func normalizeFrameworkDataSourceAttribute(a dschema.Attribute) *NormalizedAttribute {
	na := &NormalizedAttribute{
		Required:   a.IsRequired(),
		Optional:   a.IsOptional(),
		Computed:   a.IsComputed(),
		Sensitive:  a.IsSensitive(),
		Deprecated: a.GetDeprecationMessage() != "",
	}

	switch typed := a.(type) {
	case dschema.StringAttribute:
		na.Type = "string"
	case dschema.BoolAttribute:
		na.Type = "bool"
	case dschema.Int64Attribute, dschema.Float64Attribute, dschema.NumberAttribute:
		na.Type = "number"
	case dschema.ListAttribute:
		na.Type = fmt.Sprintf("list(%s)", frameworkTypeToString(typed.ElementType))
	case dschema.SetAttribute:
		na.Type = fmt.Sprintf("set(%s)", frameworkTypeToString(typed.ElementType))
	case dschema.MapAttribute:
		na.Type = fmt.Sprintf("map(%s)", frameworkTypeToString(typed.ElementType))
	case dschema.ListNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "list",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.NestedObject.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkDataSourceAttribute(child)
		}
	case dschema.SetNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "set",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.NestedObject.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkDataSourceAttribute(child)
		}
	case dschema.MapNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "map",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.NestedObject.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkDataSourceAttribute(child)
		}
	case dschema.SingleNestedAttribute:
		na.NestedType = &NormalizedNestedType{
			NestingMode: "single",
			Attributes:  make(map[string]*NormalizedAttribute),
		}
		for name, child := range typed.Attributes {
			na.NestedType.Attributes[name] = normalizeFrameworkDataSourceAttribute(child)
		}
	default:
		na.Type = CanonicalizeType(typed.GetType().String())
	}

	return na
}

func normalizeFrameworkProviderAttribute(a pschema.Attribute) *NormalizedAttribute {
	na := &NormalizedAttribute{
		Required:   a.IsRequired(),
		Optional:   a.IsOptional(),
		Computed:   a.IsComputed(),
		Sensitive:  a.IsSensitive(),
		Deprecated: a.GetDeprecationMessage() != "",
	}

	switch typed := a.(type) {
	case pschema.StringAttribute:
		na.Type = "string"
	case pschema.BoolAttribute:
		na.Type = "bool"
	case pschema.Int64Attribute, pschema.Float64Attribute, pschema.NumberAttribute:
		na.Type = "number"
	case pschema.ListAttribute:
		na.Type = fmt.Sprintf("list(%s)", frameworkTypeToString(typed.ElementType))
	case pschema.SetAttribute:
		na.Type = fmt.Sprintf("set(%s)", frameworkTypeToString(typed.ElementType))
	case pschema.MapAttribute:
		na.Type = fmt.Sprintf("map(%s)", frameworkTypeToString(typed.ElementType))
	default:
		na.Type = CanonicalizeType(typed.GetType().String())
	}

	return na
}

func normalizeFrameworkResourceBlock(b rschema.Block) *NormalizedBlockType {
	switch typed := b.(type) {
	case rschema.SingleNestedBlock:
		nb := &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
		for name, attr := range typed.Attributes {
			nb.Attributes[name] = normalizeFrameworkResourceAttribute(attr)
		}
		for name, child := range typed.Blocks {
			nb.Blocks[name] = normalizeFrameworkResourceBlock(child)
		}
		return &NormalizedBlockType{
			NestingMode: "single",
			Block:       nb,
		}
	case rschema.ListNestedBlock:
		nb := &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
		for name, attr := range typed.NestedObject.Attributes {
			nb.Attributes[name] = normalizeFrameworkResourceAttribute(attr)
		}
		for name, child := range typed.NestedObject.Blocks {
			nb.Blocks[name] = normalizeFrameworkResourceBlock(child)
		}
		return &NormalizedBlockType{
			NestingMode: "list",
			Block:       nb,
		}
	case rschema.SetNestedBlock:
		nb := &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
		for name, attr := range typed.NestedObject.Attributes {
			nb.Attributes[name] = normalizeFrameworkResourceAttribute(attr)
		}
		for name, child := range typed.NestedObject.Blocks {
			nb.Blocks[name] = normalizeFrameworkResourceBlock(child)
		}
		return &NormalizedBlockType{
			NestingMode: "set",
			Block:       nb,
		}
	default:
		return &NormalizedBlockType{
			NestingMode: "single",
			Block: &NormalizedBlock{
				Attributes: make(map[string]*NormalizedAttribute),
				Blocks:     make(map[string]*NormalizedBlockType),
			},
		}
	}
}

func normalizeFrameworkDataSourceBlock(b dschema.Block) *NormalizedBlockType {
	switch typed := b.(type) {
	case dschema.SingleNestedBlock:
		nb := &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
		for name, attr := range typed.Attributes {
			nb.Attributes[name] = normalizeFrameworkDataSourceAttribute(attr)
		}
		for name, child := range typed.Blocks {
			nb.Blocks[name] = normalizeFrameworkDataSourceBlock(child)
		}
		return &NormalizedBlockType{
			NestingMode: "single",
			Block:       nb,
		}
	case dschema.ListNestedBlock:
		nb := &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
		for name, attr := range typed.NestedObject.Attributes {
			nb.Attributes[name] = normalizeFrameworkDataSourceAttribute(attr)
		}
		for name, child := range typed.NestedObject.Blocks {
			nb.Blocks[name] = normalizeFrameworkDataSourceBlock(child)
		}
		return &NormalizedBlockType{
			NestingMode: "list",
			Block:       nb,
		}
	default:
		return &NormalizedBlockType{
			NestingMode: "single",
			Block: &NormalizedBlock{
				Attributes: make(map[string]*NormalizedAttribute),
				Blocks:     make(map[string]*NormalizedBlockType),
			},
		}
	}
}

func normalizeFrameworkProviderBlock(b pschema.Block) *NormalizedBlockType {
	switch typed := b.(type) {
	case pschema.SingleNestedBlock:
		nb := &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
		for name, attr := range typed.Attributes {
			nb.Attributes[name] = normalizeFrameworkProviderAttribute(attr)
		}
		for name, child := range typed.Blocks {
			nb.Blocks[name] = normalizeFrameworkProviderBlock(child)
		}
		return &NormalizedBlockType{
			NestingMode: "single",
			Block:       nb,
		}
	case pschema.ListNestedBlock:
		nb := &NormalizedBlock{
			Attributes: make(map[string]*NormalizedAttribute),
			Blocks:     make(map[string]*NormalizedBlockType),
		}
		for name, attr := range typed.NestedObject.Attributes {
			nb.Attributes[name] = normalizeFrameworkProviderAttribute(attr)
		}
		for name, child := range typed.NestedObject.Blocks {
			nb.Blocks[name] = normalizeFrameworkProviderBlock(child)
		}
		return &NormalizedBlockType{
			NestingMode: "list",
			Block:       nb,
		}
	default:
		return &NormalizedBlockType{
			NestingMode: "single",
			Block: &NormalizedBlock{
				Attributes: make(map[string]*NormalizedAttribute),
				Blocks:     make(map[string]*NormalizedBlockType),
			},
		}
	}
}

func frameworkTypeToString(t attr.Type) string {
	if t == nil {
		return "dynamic"
	}
	switch t {
	case types.StringType:
		return "string"
	case types.BoolType:
		return "bool"
	case types.Int64Type, types.Float64Type, types.NumberType:
		return "number"
	default:
		return CanonicalizeType(t.String())
	}
}

// ExportJSON serializes NormalizedSchema to deterministic indented JSON.
func ExportJSON(s *NormalizedSchema) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// LoadNormalizedJSON parses a NormalizedSchema from previously exported JSON.
func LoadNormalizedJSON(data []byte) (*NormalizedSchema, error) {
	var s NormalizedSchema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse normalized schema json: %w", err)
	}
	return &s, nil
}
