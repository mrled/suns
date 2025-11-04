+++
title = "TODO"
draft = true
+++

* Implement remaining symmetry types
* Deploy website
    * Handle `https://zq.suns.bz//:sdʇʇɥ` URL: redirect `suns.bz`, `zq.suns.bz` to that page
* Create tools page
    * Punycode URLs that resolve anything to a static page, to test how browsers render URLs in the bar
    * Find words that start/end/contain a substring
    * Show flips, mirrors, reverses, upside downs, etc
* Open questions
    * How do we prevent a single domain from belonging to more than one owner?
    * Do we require that the actual domain point to something, or just the TXT record?
      I think just the TXT record for now.
      Maybe in the game, add points for the DNS record to point somewhere, points for HTTPS services on it, etc.
