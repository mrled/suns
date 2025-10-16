# suns

Society for Universal Name Symmetry

## Infrastructure

- Code: GitHub
    - GitHub Actions deploys
- Deployed to AWS
    - Hugo site, deployed to S3
    - DNS, deployed to Route53
    - HTTPS by CloudFront, with content served from S3 and redirects handled by CloudFront Functions
    - Webhooks handled by Lambda
    - next/prev webring endpoints could be hardcoded at build time
    - random webring endpoint could be static and handled by client side javascript
- Future:
    - Running our own DNS servers?

## Web Servers

- Webhook receiver
    - Checks incoming domains for the required TXT records, and adds them to the webring
- Webring endpoint: next/prev/random, or something like that
- Homepage
    - Webring assets, graphics
    - List membership
    - Explanation of what SUNS is about
- Tools page
    - Punycode URLs that resolve anything to a static page, to test how browsers render URLs in the bar
    - Find words that start/end/contain a substring
    - Show flips, mirrors, reverses, upside downs, etc

## Other

- Social media?
    - Mastodon
    - Bluesky
    - Post when someone joins the network
    - Email: probably useless, but more symmetrical than Mastodon or Bluesky

## Hostname ideas

DNS Servers

- `zq.suns.sup.z.dns.suns.bz`
- `zq.suns.sup.x.dns.suns.bz`
- `zq.suns.sup.s.dns.suns.bz`
- `zq.suns.sup.o.dns.suns.bz`
- `zq.suns.sup.l.dns.suns.bz`
- `zq.suns.sup.dp.dns.suns.bz`
- `zq.suns.sup.qb.dns.suns.bz`
- `zq.suns.sup.xxx.dns.suns.bz`

Web servers

- `zq.suns.qnd.un.pub.suns.bz`
- `zq.suns.qnd.sos.pub.suns.bz`
