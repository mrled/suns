# GitHub Actions OIDC Setup

One-time setup to enable GitHub Actions to deploy CDK infrastructure using OIDC authentication.

Assumes you already have:

- A dedicated account (we are not locking down CI token permissions)
- A way to create resources in that account from the command line for this manual configuration

```sh
IAM_ROLE_NAME="SunsCiCdIamRole"

# Create the OIDC Provider.
# <https://github.blog/changelog/2023-06-27-github-actions-update-on-oidc-integration-with-aws/>
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1 1c58a3a8518e8759bf075b76b750d4f2df264fcd

# Create the IAM role for CI.
# This trust-policy.json file has hard coded the AWS account ID and GitHub org/repo/branch strings already.
aws iam create-role \
  --role-name ${IAM_ROLE_NAME} \
  --assume-role-policy-document file://trust-policy.json \
  --description "Role for GitHub Actions to deploy CDK infrastructure"

# Attach deployment policy.
# This policy gives CI the ability to CRUD necessary objects
aws iam put-role-policy \
  --role-name ${IAM_ROLE_NAME} \
  --policy-name CDKDeploymentPolicy \
  --policy-document file://cipolicy.json

# Retrieve the role ARN
aws iam get-role --role-name ${IAM_ROLE_NAME} --query 'Role.Arn' --output text
# Add the ARN to GitHub secrets (via the GitHub web UI)
# Repository Settings > Secrets and variables > Actions > New repository secret > Name "AWS_ROLE_ARN" > Value: ARN from above.
```

## Bootstrap CDK

We have to bootstrap CDK once in each region we use.
That means at least `us-east-1` because ACM certs must come from there,
and `us-east-2` because we're using that region for everything else.

```sh
cdk bootstrap aws://${AWS_ACCOUNT_ID}/us-east-1
cdk bootstrap aws://${AWS_ACCOUNT_ID}/us-east-2
```

## Updating CI permissions

When you add new CDK stacks with different AWS resources, update cipolicy.json.

Example: Adding Lambda support:

```json
{
  "Sid": "LambdaManagement",
  "Effect": "Allow",
  "Action": [
    "lambda:CreateFunction",
    "lambda:DeleteFunction",
    "lambda:GetFunction",
    "lambda:UpdateFunctionCode",
    "lambda:UpdateFunctionConfiguration",
    "lambda:TagResource"
  ],
  "Resource": "*"
}
```

Apply the updated policy:

```bash
aws iam put-role-policy \
  --role-name ${ROLE_NAME} \
  --policy-name CDKDeploymentPolicy \
  --policy-document file://cipolicy.json
```
