+++
title = "suns system design"
+++

How domain data is processed.

Conceptually, we need to track:

* Owner
* SymmetryType
* Hostname
* Group ID
* Validated At

These are stored in Dynamo DB.

The group ID is predictable and calculated independently by the client and the server,
and contains the type, and a hash of the owner plus every hostname in the group.

This is written to Dynamo DB,
and sent through Dynamo DB Streams to a Lambda worker
which builds a JSON file and uploads it to S3.

Lambda worker may just read the entire Dynamo database every time,
or read from S3 / diff just the stream change / write to S3,
depending on how much these options cost.
Streams are "at least once", so the builder must be able to tell if a record is already inserted.
The Lambda worker is not concurrent,
but this is ok be Dynamo has our backpressure.
It should write to a temp key and then PUT to the final key, because S3 overwrites are atomic and strongly consistent,
so readers won't see a partial file this way.

Nothing but the builder Lambda ever reads from Dynamo ---
the web client and the scheduled validation lambda just read from the JSON file in S3.
This means we aren't worried about primary/secondary keys in Dynamo.

## Invalid state

I want to make invalid state unrepresentable as much as possible.

Unrepresentable invalid state:

* Groups must all belong to the same owner.
  Mostly unrepresentable because the group ID contains the owner and all domains.

Representable invalid state

* Group IDs are opaque to the database and must be enforced by the application.
* A group that does not satisfy the business logic of domain validation.
  E.g. a single domain `DoubleFlip180` group.

## DNS verification

Require a special TXT record `_sums` for each domain.

* If the domain is `example.com`, look for `_suns.example.com`
* If the domain is `a.b.c.d.example.com`, look for `_suns.a.b.c.d.example.com`
* Each domain in a type must have the same TXT record set.
* Allow one CNAME hop.
  Allowing a CNAME lets users delegate control to another zone.
  Limit to one hop to keep verification deterministic.
  (Not sure if this is important?)
