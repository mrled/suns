export const config = {
  // Deployment region for most resources
  deployRegion: 'us-east-2',
  // ACM certificates must be in us-east-1 for CloudFront
  acmRegion: 'us-east-1',
  // The domain name we're using
  domainName: 'suns.bz',
  // Prefix for stack names
  stackPrefix: 'Suns',
} as const;

export const repositoryRoot = __dirname + '/../../';
