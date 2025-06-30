package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stoewer/go-strcase"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CustomResource{}

var skipAttributes = map[string]interface{}{"kind": nil, "apiVersion": nil, "status": nil}

// var _ resource.ResourceWithImportState = &CustomResource{}

func NewCustomResource(v string, g string, n v1.CustomResourceDefinitionNames, s *spec.Schema) resource.Resource {
	return &CustomResource{
		name:   resourceName(v, g, n.Singular),
		schema: s,
	}
}

// CustomResource defines the resource implementation.
type CustomResource struct {
	name   string
	schema *spec.Schema
}

func (r *CustomResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.name
}

func (r *CustomResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attr := make(map[string]schema.Attribute)
	rqat := make(map[string]bool)
	for _, r := range r.schema.Required {
		rqat[r] = true
	}
	for k, v := range r.schema.Properties {
		if _, ok := skipAttributes[k]; ok {
			continue
		}
		_, rq := rqat[k]
		av := attributeFromOAPI(&v, rq)
		if av == nil {
			continue
		}
		attr[strcase.SnakeCase(k)] = av
	}
	resp.Schema.Version = 1
	resp.Schema.Attributes = attr
}

func (r *CustomResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *CustomResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *CustomResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *CustomResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func resourceName(version string, group string, kind string) string {
	g := strings.ReplaceAll(group, ".", "_")
	return fmt.Sprintf("%s_%s_%s", g, version, kind)
}

func attributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	if s == nil {
		log.Fatal("nil input schema")
	}
	if v, ok := s.Extensions["x-kubernetes-preserve-unknown-fields"]; ok {
		bv, ok := v.(bool)
		if ok && bv {
			return dynamicAttributeFromOAPI(s, r)
		}
	}
	switch {
	case s.Type.Contains("string"):
		return stringAttributeFromOAPI(s, r)
	case s.Type.Contains("integer"):
		switch s.Format {
		case "int32":
			return int32AttributeFromOAPI(s, r)
		case "int64":
			return int64AttributeFromOAPI(s, r)
		}
	case s.Type.Contains("number"):
		switch s.Format {
		case "float":
			return floatAttributeFromOAPI(s, r)
		case "double":
			return doubleAttributeFromOAPI(s, r)
		}
	case s.Type.Contains("boolean"):
		return boolAttributeFromOAPI(s, r)
	case len(s.Type) == 0:
		log.Printf("unknown attribute type: %#v", *s)
	case s.Type.Contains("object"):
		switch {
		case len(s.Properties) > 0:
			return singleNestedAttributeFromOAPI(s, r)
		case s.AdditionalProperties.Allows && len(s.Properties) == 0:
			if isOAPIPrimitive(s.AdditionalProperties.Schema.Type) {
				return mapAttributeFromOAPI(s, r)
			} else {
				return mapNestedAttributeFromOAPI(s, r)
			}
		}
	case s.Type.Contains("array"):
		if isOAPIPrimitive(s.Items.Schema.Type) {
			return listAttributeFromOAPI(s, r)
		} else {
			return listNestedAttributeFromOAPI(s, r)
		}
	default:
		log.Printf("unsupported attribute type: %#v", s.Type)
	}
	return nil
}

func isOAPIPrimitive(t spec.StringOrArray) bool {
	switch {
	case t.Contains("string"):
		return true
	case t.Contains("number"):
		return true
	case t.Contains("integer"):
		return true
	case t.Contains("boolean"):
		return true
	default:
		return false
	}
}

func fwtypeFromOAPIPrimitive(t string, f string) attr.Type {
	switch t {
	case "string":
		return basetypes.StringType{}
	case "boolean":
		return basetypes.BoolType{}
	case "integer":
		switch f {
		case "int32":
			return basetypes.Int32Type{}
		case "int64":
			return basetypes.Int64Type{}
		}
	case "number":
		switch f {
		case "float":
			return basetypes.Float32Type{}
		case "double":
			return basetypes.Float64Type{}
		}
	}
	return nil
}

func stringAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	return schema.StringAttribute{
		Description: s.Description,
		Required:    r,
		Optional:    !r,
	}
}

func boolAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	return schema.BoolAttribute{
		Description: s.Description,
		Required:    r,
		Optional:    !r,
	}
}

func int32AttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	return schema.Int32Attribute{
		Description: s.Description,
		Required:    r,
		Optional:    !r,
	}
}

func int64AttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	return schema.Int64Attribute{
		Description: s.Description,
		Required:    r,
		Optional:    !r,
	}
}

func floatAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	return schema.Float64Attribute{
		Description: s.Description,
		Required:    r,
		Optional:    !r,
	}
}

func doubleAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	return schema.Float32Attribute{
		Description: s.Description,
		Required:    r,
		Optional:    !r,
	}
}

func dynamicAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	return schema.DynamicAttribute{
		Description: s.Description,
		Required:    r,
		Optional:    !r,
	}
}

func singleNestedAttributeFromOAPI(s *spec.Schema, r bool) schema.SingleNestedAttribute {
	att := schema.SingleNestedAttribute{
		Required:   r,
		Optional:   !r,
		Attributes: make(map[string]schema.Attribute),
	}
	rqat := make(map[string]bool)
	for _, r := range s.Required {
		rqat[r] = true
	}
	for k, p := range s.Properties {
		_, rq := rqat[k]
		av := attributeFromOAPI(&p, rq)
		if av == nil {
			continue
		}
		att.Attributes[strcase.SnakeCase(k)] = av
	}
	return att
}

func mapAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	et := fwtypeFromOAPIPrimitive(s.AdditionalProperties.Schema.Type[0], s.AdditionalProperties.Schema.Format)
	if et == nil {
		log.Fatalln("failed to determine primitive type from OpenAPI")
	}
	return schema.MapAttribute{
		Required:    r,
		Optional:    !r,
		Description: s.Description,
		ElementType: et,
	}
}

func listAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	et := fwtypeFromOAPIPrimitive(s.Items.Schema.Type[0], s.Items.Schema.Format)
	if et == nil {
		log.Fatalln("failed to determine primitive type from OpenAPI")
	}
	return schema.ListAttribute{
		Required:    r,
		Optional:    !r,
		Description: s.Description,
		ElementType: et,
	}
}

func mapNestedAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	no, ok := singleNestedAttributeFromOAPI(s.AdditionalProperties.Schema, true).GetNestedObject().(schema.NestedAttributeObject)
	if !ok {
		log.Fatalf("missmatched types - should not happen")
	}
	return schema.MapNestedAttribute{
		Required:     r,
		Optional:     !r,
		Description:  s.Description,
		NestedObject: no,
	}
}

func listNestedAttributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	no, ok := singleNestedAttributeFromOAPI(s.Items.Schema, true).GetNestedObject().(schema.NestedAttributeObject)
	if !ok {
		log.Fatalf("missmatched types - should not happen")
	}
	return schema.ListNestedAttribute{
		Required:     r,
		Optional:     !r,
		Description:  s.Description,
		NestedObject: no,
	}
}
