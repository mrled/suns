+++
title = "groupid"
+++

Group IDs contain information about the type of [symmetry]({{< ref "symmetries" >}}),
the group's owner,
and the domains in the group.

Here's an example:

```text
v1:a:DUS2oe94xFjaxf4CvZWLOyTRWJEXKgy6BtjfEXOHkwk=:+KAF43z0uQ/2zuW1oGrMaia5H6QU+3ZIRKEo2lldJzs=
```

Breaking this down, it's a string with four components separated by colon characters (`:`).

1.  The string `v1`.
2.  The symmetry type code.
    In our example, the Palindrome type code is `a`.
3.  A sha256 hash of the owner URL.
    In this example, `sha256(https://example.blog)`.
4.  A sha256 hash of all the domains in the group.
    In this example, we only have one domain in the group, so
    `sha256([etutitsni.elpmaxe.example.institute])`.

<script src="/groupid-calculator.js"></script>
<groupid-calculator></groupid-calculator>
