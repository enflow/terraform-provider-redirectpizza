# redirect.pizza Terraform Provider

Terraform provider for [redirect.pizza](https://redirect.pizza). This provides provides a single resource `redirectpizza_redirect` to manage all your redirecting needs.

We recommend using the `redirectpizza_redirect` in combination with your DNS Terraform provider is used. 
You may hardcode the values from https://redirect.pizza/support/dns-type to use for your DNS records or your own [Dedicated IP](https://redirect.pizza/support/dedicated-ip) address.

## Using the provider

```
terraform {
  required_providers {
    redirectpizza = {
      source  = "github.com/enflow/redirectpizza"
      version = "0.1.0"
    }
  }
}

# Set the variable value in *.tfvars file
# or using -var="rp_token=..." CLI option
variable "rp_token" {}

provider "redirectpizza" {
  token = var.rp_token
}

resource "redirectpizza_redirect" "old-domain" {
  sources = [
    "old-domain.com"
  ]
  destination {
    url = "new-domain.com"
  }

  # Optional
  # Must be one of:
  # - permanent
  # - permanent:307
  # - permanent:308
  # - temporary
  # - frame
  redirect_type = "permanent"

  # Optional
  tracking       = true
  uri_forwarding = true
  keep_query_string     = false
  tags                  = ["prod", "dev"]
}
```

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.18

## Import
To import existing redirect.pizza redirects to Terraform's state, you may use the `import` block. 
The `id` can be found in the redirect.pizza's API response or in the URL when navigating to a specific redirect. Example: https://redirect.pizza/redirects/123123123

```
terraform import redirectpizza_redirect.old-domain.com 123123123
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

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

To run locally, run `(cd ..; export GOBIN=$(pwd); go install) && TF_LOG=WARN TF_VAR_rp_token=rpa_xxxxxxxxxxxx terraform apply` in the `examples/` directory.
