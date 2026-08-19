# redirect.pizza Terraform Provider

Terraform provider for [redirect.pizza](https://redirect.pizza). Use the `redirectpizza_redirect` resource to manage redirects as code.

Use this provider together with your DNS provider. You may hardcode values from the [DNS records](https://redirect.pizza/support/dns-type) documentation, or your own [Dedicated IP](https://redirect.pizza/support/dedicated-ip).

## Using the provider

```hcl
terraform {
  required_providers {
    redirectpizza = {
      source  = "enflow/redirectpizza"
      version = "~> 0.2"
    }
  }
}

# Set the variable value in a *.tfvars file or with -var="rp_token=..."
variable "rp_token" {
  type      = string
  sensitive = true
}

provider "redirectpizza" {
  token = var.rp_token
}

resource "redirectpizza_redirect" "old-domain" {
  sources = [
    "old-domain.com",
  ]

  destination {
    url = "https://new-domain.com"
  }

  # Optional. One of: permanent, temporary, permanent:308, temporary:307, frame
  redirect_type = "permanent"

  tracking          = true
  uri_forwarding    = true
  keep_query_string = false
  tags              = ["prod", "dev"]
}
```

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18

## Import

The import ID is the numeric redirect ID from the API or from the dashboard URL (for example `https://redirect.pizza/redirects/123123123`).

Terraform 1.5+ `import` block:

```hcl
import {
  to = redirectpizza_redirect.old-domain
  id = "123123123"
}
```

CLI:

```sh
terraform import redirectpizza_redirect.old-domain 123123123
```

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```sh
$ go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to this Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

To run locally, add a [development override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) for `enflow/redirectpizza` (or `registry.terraform.io/enflow/redirectpizza`) pointing at the directory that contains the built binary. Then from `examples/`:

```sh
(cd ..; export GOBIN=$(pwd); go install) && TF_LOG=WARN TF_VAR_rp_token=rpa_xxxxxxxxxxxx terraform apply
```
