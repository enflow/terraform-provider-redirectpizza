resource "redirectpizza_redirect" "old-source" {
  sources = [
    "old-source.be",
    "old-source.uk",
    "old-source.nl",
  ]
  destination {
    url = "new-fancy-site3.nl"
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
  tracking          = true
  uri_forwarding    = true
  keep_query_string = false
  tags              = ["prod", "dev"]
}

# Dynamic destinations
resource "redirectpizza_redirect" "old-source2" {
  sources = [
    "old-dynamic-source.com",
  ]

  destination {
    url        = "new-fancy-site.nl"
    expression = "country == 'NL'"

    # Optional. Must be one of: inherit (default), enabled, disabled
    monitoring = "enabled"
  }
  destination {
    url        = "new-fancy-site.us"
    expression = "country == 'US'"
    monitoring = "disabled"
  }
  destination {
    url = "new-fancy-site.com"
  }
}