# suns project AGENTS.md

* AWS CDK is in `/aws`
* A Go project with a command line for local use, Lambdas, etc is in `/symval`
* A Hugo static website is in `/www`

## Interacting with AWS

Never try to use the AWS API yourself, just tell the user if you need to deploy with CDK or run the `aws` commandline.

Feel free to use `curl` to talk to the endpoint we're building in AWS, though. It doesn't require authentication.

## Interacting with the repository

Do not make git commits unless specifically instructed to do so.

Feel free to use git commands to explore repository history etc as required.
