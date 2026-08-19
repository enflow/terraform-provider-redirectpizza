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
