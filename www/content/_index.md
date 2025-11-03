+++
title = "Society for Universal Name Symmetry"
+++

The Society for Universal Name Symmetry is a club open to anyone with a symmetric DNS name.

DNS name symmetry can be achieved in several ways.
Some examples:

- Palindrome: `zb.snus.suns.bz`
    - Bonus for a fully palindrom'ed URL: `https://zb.snus.suns.bz//:sptth`
- Single 180° flip: `zq.suns.bz` (`zq.su` + `ns.bz`, flip either half 180° to get the other half)
    - Bonus for a fully flipped URL: `https://zq.suns.bz//:sdʇʇɥ`
- Double 180° flip: `zq.su` / `ns.bz` (example domains that we don't own)
- Mirrored text: `duq.xodbox.pub` (example domain that we don't own)
- Mirrored DNS names: `me.example.com` / `com.example.me`
- Antonymmic DNS names: `https://at.example.email` / `https@example.website`

## Joining suns

Want to join?

- Create a symmetrical name with one of the methods above
- Create TXT records for *each* domain
- POST to `https://zq.suns.bz/api/v1/attest`

Any domain owner can join by creating a palindrome of their domain.
For instance, if you own `example.institute`,
create a DNS record for `etutitsni.elpmaxe.example.institute`.

This works fine even for subdomains.
If you control DNS for a subdomain like `example.com.us`,
create a DNS record for `su.moc.elpmaxe.example.com.us`.

Of course, you can also join with one of the other methods above, like flips or mirrors.

Each record has an owner, which we expect to be a URL,
and which is rendered clickable in the members table below.

Here's an example submission with curl:

```sh
curl -X POST https://zq.suns.bz/api/v1/attest \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "https://example.blog",
    "type": "palindrome",
    "domains": ["etutitsni.elpmaxe.example.institute"]
  }'
```

The owner can be any URL and doesn't have to match one of the entries in the `domains` list;
`https://example.blog` might own `example.institute`.
Also note that the `owner` typically includes an `https://` prefix,
but the `domains` list should be bare DNS names.


## Members

<script src="/domain-records.js"></script>
<domain-records src='{{< param "recordsDomainsPath" >}}' priority-owner="https://zq.suns.bz"></domain-records>
