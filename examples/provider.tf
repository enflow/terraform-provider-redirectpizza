terraform {
  required_providers {
    redirectpizza = {
      source = "github.com/enflow/redirectpizza"
    }
  }
}

provider "redirectpizza" {
  # token = "FILL ME IN"

  # Optional, no need to specify:
  # api_base_url = "https://redirect.pizza/api/"
}
