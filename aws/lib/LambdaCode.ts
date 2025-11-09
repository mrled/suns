import * as lambda from "aws-cdk-lib/aws-lambda";
import { repositoryRoot } from "./config";
import * as path from "path";
import { Construct } from "constructs";

/**
 * Returns Lambda code configuration for the unified Lambda binary.
 * This should be called once per Lambda function. CDK will cache the build output
 * and reuse it across Lambdas that use the same source path and bundling configuration.
 * Each Lambda will use the LAMBDA_HANDLER environment variable to determine which handler to run.
 */
export function getUnifiedLambdaCode(scope: Construct): lambda.AssetCode {
  return lambda.Code.fromAsset(path.join(repositoryRoot, "symval"), {
    bundling: {
      image: lambda.Runtime.PROVIDED_AL2023.bundlingImage,
      command: [
        "sh",
        "-c",
        [
          "dnf install -y golang",
          "export GOOS=linux",
          "export GOARCH=arm64",
          "export CGO_ENABLED=0",
          "cd /asset-input",
          'go build -tags netgo -ldflags "-s -w -extldflags -static" -trimpath -o /asset-output/bootstrap ./cmd/lambda',
        ].join(" && "),
      ],
      user: "root",
    },
  });
}
