export const config = {
  // Deployment region for most resources
  deployRegion: "us-east-2",
  // ACM certificates must be in us-east-1 for CloudFront
  acmRegion: "us-east-1",
  // The domain name we're using
  domainName: "suns.bz",
  // Prefix for stack names
  stackPrefix: "Suns",
  // S3 key for the domains JSON file
  domainsDataKey: "records/domains.json",
  // Where alerts are sent
  alertEmail: "me+suns-alerts@micahrl.com",
  // Lambda function names - centralized to avoid cross-stack reference issues
  // If we used generated names, we wouldn't be able to ever deploy a new version of a function,
  // because other stacks would depend on the old name (and the generated name changes with each deployment).
  functionNames: {
    httpApi: "SunsHttpApiFunction",
    streamer: "SunsStreamerFunction",
    reattestBatch: "SunsReattestBatchFunction",
  },
} as const;

export const repositoryRoot = __dirname + "/../../";
