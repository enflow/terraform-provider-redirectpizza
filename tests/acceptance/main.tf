terraform {
  required_providers {
    redirectpizza = {
      source  = "github.com/enflow/redirectpizza"
    }
  }
}

provider "redirectpizza" { }

resource "redirectpizza_redirect" "testing-domain" {
  sources = [
    "redirect-pizza-terraform-testing-domain.com"
  ]
  destination {
    url = "redirect.pizza"
  }
}
