resource "redirectpizza_redirect" "example" {
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
  tags              = ["prod", "legacy"]
}

# Dynamic destinations: all but the fallback need an expression.
resource "redirectpizza_redirect" "geo" {
  sources = [
    "old-dynamic-source.com",
  ]

  destination {
    url        = "https://new-fancy-site.nl"
    expression = "country == 'NL'"
    monitoring = "enabled"
  }

  destination {
    url        = "https://new-fancy-site.us"
    expression = "country == 'US'"
    monitoring = "disabled"
  }

  destination {
    url = "https://new-fancy-site.com"
  }
}
