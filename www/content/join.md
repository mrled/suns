+++
title = "Joining suns"
+++

Membership is open to anyone with control of a DNS zone.

1.  Create one or more symmetrical DNS names from the [list]({{< ref "symmetries" >}}) of symmetries
2.  Decide on a URL to use as your owner ID
3.  Calculate your symmetry's Group ID
4.  Create TXT records for each domain in the group
5.  POST to `https://zq.suns.bz/api/v1/attest`

For example:

1.  The owner of `example.institute` wants to set up Palindrome symmetry,
    so they create a regular DNS record (`A`, `AAAA`, `CNAME`, however they want to use it)
    for `etutitsni.elpmaxe.example.institute`.

2.  They use `https://example.blog` as their main website,
    so they use that as the owner ID.
    The owner URL does not have to be one of the domains in the group.
    It should have an `https://` prefix,
    and it may contain a path like `https://example.blog/about-me` if you like.

3.  They calculate the group ID as
    `v1:a:DUS2oe94xFjaxf4CvZWLOyTRWJEXKgy6BtjfEXOHkwk=:+KAF43z0uQ/2zuW1oGrMaia5H6QU+3ZIRKEo2lldJzs=`.
    See ths [groupid]({{< ref "groupid" >}}) page for an explanation of this value and a calculator.

4.  They create a TXT record to attest ownership for every domain in the group with the group ID.
    In this case, that means a record at `_suns.etutitsni.elpmaxe.example.institute`
    that contains the group ID.

5.  POST to the API endpoint, like this:

    ```sh
    curl -X POST https://zq.suns.bz/api/v1/attest \
      -H "Content-Type: application/json" \
      -d '{
        "owner": "https://example.blog",
        "type": "palindrome",
        "domains": ["etutitsni.elpmaxe.example.institute"]
      }'
    ```

Membership remains valid as long as the attestation records stay in place.