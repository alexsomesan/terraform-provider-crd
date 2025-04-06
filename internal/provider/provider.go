// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-codegen-openapi/pkg/config"
	"github.com/hashicorp/terraform-plugin-codegen-openapi/pkg/mapper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pb33f/libopenapi"

	"github.com/hashicorp/terraform-plugin-codegen-openapi/pkg/explorer"
)

// Ensure KubernetesCRD satisfies various provider interfaces.
var _ provider.Provider = &KubernetesCRD{}
var _ provider.ProviderWithFunctions = &KubernetesCRD{}

// KubernetesCRD defines the provider implementation.
type KubernetesCRD struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	clients *KubernetesClients
}

// KubernetesCRDModel describes the provider data model.
type KubernetesCRDModel struct {
	Kubeconfig types.String `tfsdk:"kubeconfig"`
}

func (p *KubernetesCRD) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "crd"
	resp.Version = p.version
}

func (p *KubernetesCRD) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	// spec := map[string]*spec.Schema{}

	paths, err := p.clients.discovery.OpenAPIV3().Paths()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get OpenAPI paths", err.Error())
		return
	}

	for path, gv := range paths {
		ks := strings.Split(path, "/")
		if ks[0] != "apis" {
			continue
		}
		s, err := gv.Schema("application/json")
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to decode schema for resource %q", strings.Join(ks[1:], "/")), err.Error())
			continue
		}

		// create a new document from specification bytes
		document, err := libopenapi.NewDocument(s)
		// if anything went wrong, an error is thrown
		if err != nil {
			panic(fmt.Sprintf("cannot create new document: %e", err))
		}

		// because we know this is a v3 spec, we can build a ready to go model from it.
		v3Model, errors := document.BuildV3Model()
		if len(errors) > 0 {
			for i := range errors {
				fmt.Printf("error: %e\n", errors[i])
			}
			panic(fmt.Sprintf("cannot create v3 model from document: %d errors reported", len(errors)))
		}

		cfg := config.Config{
			Resources: map[string]config.Resource{},
		}

		// get a count of the number of paths and schemas.
		m := v3Model.Model
		m.Index.BuildIndex()
		schemas := m.Components.Schemas

		for schema := schemas.First(); schema != nil; schema = schema.Next() {
			// get the name of the schema
			schemaName := schema.Key()

			if rejectPath(schemaName) {
				// fmt.Printf("Skipping schema %q\n", schemaName)
				continue
			}

			// get the schema object from the map
			schemaValue := schema.Value()

			// build the schema
			schema := schemaValue.Schema()

			// if the schema has properties, print the number of properties
			if schema != nil && schema.Properties != nil {
				// oaschema, err := ogen.BuildSchema(schemaValue, ogen.SchemaOpts{}, ogen.GlobalSchemaOpts{})
				// if err != nil {
				// 	panic(fmt.Sprintf("Failed to convert OAPI schema of %q: %s", schemaName, err))
				// }
				// fmt.Printf("Schema %q is type %s\n", schemaName, oaschema.Type)

				cfg.Resources[schemaName] = config.Resource{
					Read: &config.OpenApiSpecLocation{
						Path: "",
					},
				}
			}
		}

		ex := explorer.NewConfigExplorer(v3Model.Model, cfg)
		explorerResources, err := ex.FindResources()
		if err != nil {
			panic(fmt.Sprintf("Failed to find resources in OAPI doc %q: %s", v3Model.Model.Info, err))
		}

		resourceMapper := mapper.NewResourceMapper(explorerResources, cfg)
		resourcesIR, err := resourceMapper.MapToIR(nil)
		if err != nil {
			panic(fmt.Sprintf("Failed to map resources to IR: %s", err))
		}

		for _, rir := range resourcesIR {
			fmt.Printf("Schema %q is type %s\n", rir.Name, *rir.Schema.Description)
		}

	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"kubeconfig": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
		},
	}
}

func (p *KubernetesCRD) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data KubernetesCRDModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Kubeconfig.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	resp.DataSourceData = p.clients
	resp.ResourceData = p.clients
}

func (p *KubernetesCRD) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *KubernetesCRD) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
	}
}

func (p *KubernetesCRD) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KubernetesCRD{
			version: version,
			clients: NewKubernetesClient(),
		}
	}
}
