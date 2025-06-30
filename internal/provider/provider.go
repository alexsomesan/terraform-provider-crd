// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rtschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kube-openapi/pkg/validation/spec"
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
	var resources []func() resource.Resource

	crds, err := p.clients.APIextensions.ApiextensionsV1().CustomResourceDefinitions().List(ctx, v1.ListOptions{})
	if err != nil {
		log.Fatalf("failed to list Custom Resources: %s", err)
	}

	for _, crd := range crds.Items {
		for _, ver := range crd.Spec.Versions {
			gv := rtschema.GroupVersion{Version: ver.Name, Group: crd.Spec.Group}
			gvspec, err := p.clients.Openapi.GVSpec(gv)
			if err != nil {
				log.Fatal(err)
			}
			var s *spec.Schema
			for k := range gvspec.Components.Schemas {
				if !strings.HasSuffix(k, crd.Spec.Names.Kind) {
					continue
				}
				s = gvspec.Components.Schemas[k]
				break
			}
			resources = append(resources, func() resource.Resource {
				r := NewCustomResource(ver.Name, crd.Spec.Group, crd.Spec.Names, s)
				return r
			})
		}
	}

	return resources
}

func (p *KubernetesCRD) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *KubernetesCRD) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KubernetesCRD{
			version: version,
			clients: NewKubernetesClient(),
		}
	}
}
