+++
title = "SUNS"
+++

# Society for Universal Name Symmetry

The Society for Universal Name Symmetry is a club open to anyone with a symmetric DNS name.

DNS name symmetry can be achieved in several ways.
Some examples:

- Palindrome: `zb.snus.suns.bz`
    - Bonus for a fully pallindromed URL: `https://zb.snus.suns.bz//:sptth`
- Single 180° flip: `zq.suns.bz` (`zq.su` + `ns.bz`, flip either half 180° to get the other half)
    - Bonus for a fully flipped URL: `https://zq.suns.bz//:sdʇʇɥ`
- Double 180° flip: `zq.su` / `ns.bz` (example domains that we don't own)
- Mirrored text: `duq.xodbox.pub` (example domain that we don't own)
- Mirrored DNS names: `me.example.com` / `com.example.me`
- Antonymmic DNS names: `https://at.example.email` / `https@example.website`

## Joining SUNS

Want to join?

1. Create a symmetrical name with one of the methods above
2. Create TXT records for *each* domain
3. Submit to `/webhook?type=TYPE&name=NAME` for single name symmetries
   or `/webhook?type=TYPE&name=NAME&mirror=MIRROR` for dual-name symmetries.

Any domain owner can join by creating a palindrome of their domain.
For instance, if you own `example.institute`,
create a DNS record for `etutitsni.elpmaxe.example.instutite`.

This works fine even for subdomains.
If you control DNS for a subdomain like `example.com.us`,
create a DNS record for `su.moc.elpmaxe.example.com.us`.

Of course, you can also join with one of the other methods above, like flips or mirrors.

## Members

??
