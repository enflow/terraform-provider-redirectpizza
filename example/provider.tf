terraform {
  required_providers {
    redirectpizza = {
      source  = "github.com/enflow/redirectpizza"
      version = "0.0.1"
    }
  }
}

provider "redirectpizza" {
  token = "FILL ME IN"
}
