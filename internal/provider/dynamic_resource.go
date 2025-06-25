package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stoewer/go-strcase"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CustomResource{}

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
	resp.Schema.Version = 1
	resp.Schema.Attributes = attributesFromOAPIObject(r.schema)
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
	g := strings.Replace(group, ".", "_", -1)
	return fmt.Sprintf("%s_%s_%s", g, version, kind)
}

func attributesFromOAPIObject(s *spec.Schema) map[string]schema.Attribute {
	if !s.Type.Contains("object") {
		log.Fatalf("wrong schema type at top level: %s", strings.Join(s.Type, ","))
	}
	attr := make(map[string]schema.Attribute)
	req := make(map[string]bool)
	for _, r := range s.Required {
		req[r] = true
	}
	for k, v := range s.Properties {
		_, rq := req[k]
		av := attributeFromOAPI(&v, rq)
		if av == nil {
			continue
		}
		attr[strcase.SnakeCase(k)] = av
	}
	return attr
}

func attributeFromOAPI(s *spec.Schema, r bool) schema.Attribute {
	if s == nil {
		log.Fatal("nil input schema")
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
		if v, ok := s.Extensions["x-kubernetes-preserve-unknown-fields"]; ok {
			bv, err := strconv.ParseBool(v.(string))
			if err != nil {
				log.Fatalf("failed to parse boolean value: %s", err)
			}
			if bv {
				return dynamicAttributeFromOAPI(s, r)
			}
		}
		log.Println(s.Ref)
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
