# redirect.pizza Terraform Provider

XXXXXX

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.18

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

## Using the provider

```
terraform {
  required_providers {
    redirectpizza = {
      source  = "github.com/enflow/redirectpizza"
    }
  }
}

provider "redirectpizza" {
  token = "rpa_XXXXXXXXXXXXX" # Or set the `REDIRECTPIZZA_API_TOKEN` environment variable.
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

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
