## 0.3.0 (August 19, 2026)

ENHANCEMENTS:

* Retry redirect.pizza API requests on HTTP 429, honoring `Retry-After` with jitter so parallel applies back off instead of failing immediately.

NOTES:

* Examples, README, and generated Registry docs now use the published source address `enflow/redirectpizza`.
* Resource docs include multiple destinations, nested `expression` / `monitoring` attributes, and import examples.
* Provider and resource schema descriptions are now populated so the Registry docs are complete.
