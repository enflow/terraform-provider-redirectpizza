resource "redirectpizza_redirect" "old-source" {
  sources = [
    "old-source.be",
    "old-source.uk",
    "old-source.nl",
    "old-source.co.uk",
  ]
  destination {
    url = "new-fancy-site.nl"
  }

  // Must be one of:
  // - permanent
  // - permanent:307
  // - permanent:308
  // - temporary
  // - frame
  redirect_type = "permanent"

  # Optional
  enable_tracking       = true
  enable_uri_forwarding = false
  keep_query_string     = false
  tags                  = ["prod", "test", "dev"]
}
