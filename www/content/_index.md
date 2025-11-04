+++
title = "Society for Universal Name Symmetry"
+++

The Society for Universal Name Symmetry is a club open to anyone with a symmetric DNS name.

DNS name symmetry can be achieved in several ways.
For instance, this domain name is `zq.suns.bz` which is a Single 180° Flip,
because `zq.su` is `ns.bz` flipped 180 degrees.
See the [symmetries]({{< ref "symmetries" >}}) page for a list of what we support with more examples.

## Joining suns

Membership is open to anyone with control of a DNS zone.

1.  Create one or more symmetrical DNS names from the [list]({{< ref "symmetries" >}}) of symmetries
2.  Decide on a URL to use as your owner ID
3.  Calculate your symmetry's Group ID
4.  Create TXT records for each domain in the group
5.  POST to `https://zq.suns.bz/api/v1/attest`

For example:

1.  The owner of `example.institute` wants to set up Palindrome symmetry,
    so they create a DNS record for `etutitsni.elpmaxe.example.institute`.

2.  They use `https://example.blog` as their main website,
    so they use that as the owner ID.
    The owner URL does not have to be one of the domains in the group.
    It should have an `https://` prefix,
    and it may contain a path like `https://example.blog/about-me` if you like.

3.  They calculate the Group ID as
    `v1:a:EAaArVRs5qV39C9S3zO0z9ynVoWeZkuNfeMpsVDQnOk=:UIfveywMs1sKCW+ywVfEYhuDl+s6r6H3fghgBqGwbh8=`.
    Breaking this, down, it's a string with four components separated by `:`

    1.  The string `v1`.
    2.  The symmetry type code.
        In our example, the Palindrome type code is `a`.
    3.  A sha256 hash of the owner URL.
        In this example, `sha256(https://example.blog)`.
    4.  A sha256 hash of all the domains in the group.
        In this example, we only have one domain in the group, so
        `sha256([etutitsni.elpmaxe.example.institute])`.

4.  They create a TXT record for every domain in the group with the Group ID.
    In this case, that means a record at `_suns.etutitsni.elpmaxe.example.institute`
    that contains `v1:a:EAaArVRs5qV39C9S3zO0z9ynVoWeZkuNfeMpsVDQnOk=:UIfveywMs1sKCW+ywVfEYhuDl+s6r6H3fghgBqGwbh8=`.

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

## Members

<script src="/domain-records.js"></script>
<domain-records src='{{< param "recordsDomainsPath" >}}' priority-owner="https://zq.suns.bz"></domain-records>
