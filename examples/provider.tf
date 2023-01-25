terraform {
  required_providers {
    redirectpizza = {
      source = "github.com/enflow/redirectpizza"
      version = "0.1.0"
    }
  }
}

# Set the variable value in *.tfvars file
# or using -var="rp_token=..." CLI option
variable "rp_token" {}

provider "redirectpizza" {
  token = var.rp_token

  # Optional, no need to specify:
  # api_base_url = "https://redirect.pizza/api/"
}
