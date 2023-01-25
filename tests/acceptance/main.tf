terraform {
  required_providers {
    redirectpizza = {
      source  = "github.com/enflow/redirectpizza"
    }
  }
}

# Set the variable value in *.tfvars file
# or using -var="rp_api_token=..." CLI option
variable "rp_api_token" {}

provider "redirectpizza" {
  token = var.rp_api_token
}

resource "redirectpizza_redirect" "testing-domain" {
  sources = [
    "redirect-pizza-terraform-testing-domain.com"
  ]
  destination {
    url = "redirect.pizza"
  }
}
