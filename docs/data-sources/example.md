---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "crd_example Data Source - crd"
subcategory: ""
description: |-
  Example data source
---

# crd_example (Data Source)

Example data source

## Example Usage

```terraform
data "crd_example" "example" {
  configurable_attribute = "some-value"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `configurable_attribute` (String) Example configurable attribute

### Read-Only

- `id` (String) Example identifier
