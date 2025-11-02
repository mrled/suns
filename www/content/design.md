+++
title = "suns system design"
+++

How domain data is processed.

## Group ID

When a domain is singularly symmetrical (like `a` for palindrome),
it's in a group with just itself.
When a domain is symmetrical with another domain (like `e` for mirrornames),
it's in a group with another domain.

In either case, we create a `group ID` that incorporates:

* The owner: an arbitrary string, I recommend a URL
* The symmetry type: `a` (palindrome), `e` (mirrornames), etc
* Each domain name, sorted alphabetically: `me.example.com`, `com.example.me`, etc

The algorithm for this is defined in `CalculateV1()` in `symval/internal/service/groupid/groupid.go`,

```go
// CalculateV1 generates a group ID by hashing owner + all hostnames
// The result is formatted as: idversion:type:base64(sha256(owner)):base64(sha256(sort(hostnames))).
func (s *Service) CalculateV1(owner, gtype string, hostnames []string) (string, error) {
  // ...
}
```

Implementation notes:

* In theory, this could be used to support groups of more than two domain names,
  though it's not clear what it would mean for a group of 3 or more names to be symmetrical.
  Perhaps theoretical mathematics has an answer for this important question.
* Sorting the domain names ensures that a group ID is consistent no matter who calcualtes it ---
  and ensures that a logical set of hosts cannot be counted more than once.

## Dynamo design

Conceptually, we need to track:

* Owner
* SymmetryType
* Hostname
* Group ID (see above)
* Validated At

These are stored in Dynamo DB.

## Data access patterns

* Present the data to web visitors, one of two options:
  * Present a list of all owners, to show a membership list
  * Present a list of all of an owner's domain groups, to have a page specific to each owner
* Bulk processing
  * Read through all records and validate them, deleting invalid ones

We plan to show all owners on the homepage with a list of all their domain groups.
This means that every web request for the homepage will read _every key in the table_.

Later we might paginate this, or show owners' domain groups on owner-specific pages.

## Data storage implementation

Because we need to read every key in the table,
it's more efficient to update a JSON file in S3 on every change to Dynamo
than it is to read data out of Dynamo more than once.

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
This just uses Dynamo as a way to do concurrent writes,
sort of like a queue.

Nothing but the builder Lambda ever reads from Dynamo ---
the web client and the scheduled validation lambda just read from the JSON file in S3.
This means we aren't worried about primary/secondary keys in Dynamo.

We make they Dynamo PK (partition key) a composite of: **owner + hostname + type + group ID**.
We could make a new PK concatenating that information with no SK (sorting key).
But the group ID already incorporates a hash of the owner and the type,
so we can simplify this by using the group ID as the PK and the hostname as the SK.

* Very simple schema.
* No GSIs and no mirror keys (mirrors = the same data with a different PK).
* Fits perfectly for streams or full table scans only.
* This would be terrible for querying individual records, but we aren't doing that.
* Doesn't allow for key updates; you have to delete+put because the PK is changing.
  This is ok, because the only changes we make are to update the validation time (not part of PK)
  or delete a record as invalid.
  * If the owner wants to change their URL, they'll set new DNS records, and submit,
    and the old records will be invalidated on the next scan.

## Concurrency

User requests go through the webhook Lambda which saves data to DynamoDB.

DynamoDB streams events to a streamer Lambda which saves each changed record to a JSON file in S3.
This Lambda has `reservedConcurrentExecutions: 1` to only allow one to run at a time,
which acts as a lock on writes to the JSON file.
This Lambda is the only writer to the JSON file.

There is a scheduler Lambda that is run every day that re-attests every record in the JSON file,
updating Dynamo with new validation time or deleting recordds that fail attestation.
(There is a grace period to prevent intermittent errors from removing actually valid records.)

Aside from the Lambdas, the browser retrieves the JSON file when a user visits the website.

We use a monotonic `Rev` field in our data model to prevent concurrency bugs.
When we are making changes to a record in Dynamo,
we use `ConditionExpression: rev = :snapshotRev` to ensure that
the change fails if an update has been made to Dynamo since our last snapshot of the table.

## Dynamo storage costs

* Data storage costs
    * $0.25/GB
    * Each record is about 250B
    * e.g. 4M records would be 1GB and $0.25/mo; 1k records $0.01/mo
* Webhook adding/updating new records
    * $0.000000625/write
    * e.g. with 1000 new members per day, $0.02
* Bulk process writing to all records:
    * Update or delete every record once per day
    * 30 days/mo * $0.000000625 * X records/day
    * e.g. with a constant 1000 members with avg 5 domains each (updating 5000 records every day, 150,000 records every month), $0.09/mo
* Reading all records once per day for bulk processing
    * Read full table every day
    * $0.000000125/read
    * e.g. with a constant 1000 members with avg 5 domains each,
      (reading 5000 records every day, 150,000 records every month), $0.01/mo
* Web reading records
    * Assume the member list with all domains each member owns is on the homepage: a FULL TABLE READ for every visitor
    * e.g. with 10x daily visitors as members:
      10,000 visitors x 30 days = 300,000 requests,
      each request 5000 records for 1,500,000,000 records/mo: $6.25/mo
    * $0.000000125/read
* The costs don't matter except for reading Dynamo data directly to the web. The smoothest transition is with streaming dynamo->lambda->s3 export.

## Invalid state

I want to make invalid state unrepresentable as much as possible.

Unrepresentable invalid state:

* Groups must all belong to the same owner.
  Mostly unrepresentable because the group ID contains the owner and all domains.

Representable invalid state

* Group IDs are opaque to the database and must be enforced by the application.
* A group that does not satisfy the business logic of domain validation.
  E.g. a single domain `DoubleFlip180` group.

## DNS claims

Require a special TXT record `_suns` for each domain.

* If the domain is `example.com`, look for `_suns.example.com`
* If the domain is `a.b.c.d.example.com`, look for `_suns.a.b.c.d.example.com`
* Each domain in a group must have the same TXT record set.
* The contents of the TXT record is the group ID (see above).
* Allow one CNAME hop.
  Allowing a CNAME lets users delegate control to another zone.
  Limit to one hop to keep verification deterministic.
  (Not sure if this is important?)
* Expect one TXT record for every group that the domain is in.

These records provide _claims_ that the domain is part of the group,
but they don't _verify_ the claims.

## Consitency checking

Consistency checking confirms that

* All domains in the group have valid DNS claims.
* All domains in the group have a claim record with the group ID.
* All DNS claims for all domains the group have the same owner.
  * This is subtle:
    for every domain in the group,
    check that every claim for that domain has the same owner,
    even claims for _other groups_.
    Domains can be in more than one group,
    but cannot be owned by more than one owner.

## Group validation

Group validation takes DNS claims which have been individually looked up and consistency checked,
and validates that they make sense as a whole.

* All domains are present in the group
* The group doesn't contain any spurious domains

## Attestation

Attestation ties all of these together.

* Receive input from users
* Query DNS for claims
* Consistency-check the claims
* Validate the claims

Basically a "Lookup + Validate" flow.
