package provider

// import (
// 	"net/http"

// 	"github.com/hashicorp/terraform-plugin-framework/diag"
// 	"github.com/hashicorp/terraform-plugin-framework/resource"
// 	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
// 	"k8s.io/kube-openapi/pkg/validation/spec"
// )

// // Ensure provider defined types fully satisfy framework interfaces.
// var _ resource.Resource = &DynamicResource{}
// // var _ resource.ResourceWithImportState = &DynamicResource{}

// func NewDynamicResource() resource.Resource {
// 	return &DynamicResource{}
// }

// // DynamicResource defines the resource implementation.
// type DynamicResource struct {
// 	client *http.Client
// }

// func oAPItoFrameworkSchema(o *spec.Schema) (s schema.Schema, d diag.Diagnostics) {

// }
