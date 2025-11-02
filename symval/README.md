# symval

`symval` is the SYMmetric VALidator.

Binaries:

* `httpapi`: A lambda function, run from an HTTP API, to validate whether domains are symmetric, and add them to the database
* `reattestbatch`: A lambda function, run on a schedule, to re-validate domains in the database
* `symval`: A command line function for running validation code locally

The Lambda code should _avoid cgo_ in order to reduce Lambda package size and cold start time.
