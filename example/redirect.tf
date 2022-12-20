
resource "redirectpizza_redirect" "old-source" {
  sources = [
    "old-source.nl",
    "old-source.be",
    "old-source.de",
  ]
  destination {
    url = "new-fancy-site.nl"
  }

  redirect_type     = "permanent"

# Optional:
#  keep_query_string = true
#  tracking          = true
#  tags              = [ "prod", "test", "dev" ]
}
