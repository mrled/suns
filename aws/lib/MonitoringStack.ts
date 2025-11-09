import * as cdk from "aws-cdk-lib";
import * as cloudwatch from "aws-cdk-lib/aws-cloudwatch";
import * as cloudwatchActions from "aws-cdk-lib/aws-cloudwatch-actions";
import * as sns from "aws-cdk-lib/aws-sns";
import * as snsSubscriptions from "aws-cdk-lib/aws-sns-subscriptions";
import * as logs from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";
import { config } from "./config";

export interface MonitoringStackProps extends cdk.StackProps {
  apiFunctionName: string;
  streamerFunctionName: string;
  reattestBatchFunctionName: string;
}

export class MonitoringStack extends cdk.Stack {
  public readonly alertTopic: sns.Topic;

  constructor(scope: Construct, id: string, props: MonitoringStackProps) {
    super(scope, id, props);

    // Create SNS topic for alerts
    this.alertTopic = new sns.Topic(this, "AlertTopic", {
      topicName: `${config.stackPrefix}AlertTopic`,
      displayName: `${config.stackPrefix} System Alerts`,
    });

    // Add email subscription
    this.alertTopic.addSubscription(
      new snsSubscriptions.EmailSubscription(config.alertEmail),
    );

    // Create SNS alarm action
    const alarmAction = new cloudwatchActions.SnsAction(this.alertTopic);

    // Iterator Age Alarm for DynamoDB Streams
    // This monitors how far behind the Lambda function is in processing the stream
    const iteratorAgeAlarm = new cloudwatch.Alarm(this, "IteratorAgeAlarm", {
      alarmName: `${config.stackPrefix}-StreamIteratorAge`,
      alarmDescription:
        "Alert when DynamoDB Stream processing is falling behind",
      metric: new cloudwatch.Metric({
        namespace: "AWS/Lambda",
        metricName: "IteratorAge",
        dimensionsMap: {
          FunctionName: props.streamerFunctionName,
        },
        statistic: "Maximum",
        period: cdk.Duration.minutes(1),
      }),
      threshold: 120_000, // 120 seconds in milliseconds
      evaluationPeriods: 5,
      treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
      comparisonOperator: cloudwatch.ComparisonOperator.GREATER_THAN_THRESHOLD,
    });
    iteratorAgeAlarm.addAlarmAction(alarmAction);

    // Lambda Invocation Errors Alarm for HTTP API Function
    const apiErrorAlarm = new cloudwatch.Alarm(this, "ApiErrorAlarm", {
      alarmName: `${config.stackPrefix}-ApiInvocationErrors`,
      alarmDescription: "Alert on HTTP API Lambda function invocation errors",
      metric: new cloudwatch.Metric({
        namespace: "AWS/Lambda",
        metricName: "Errors",
        dimensionsMap: {
          FunctionName: props.apiFunctionName,
        },
        statistic: "Sum",
        period: cdk.Duration.minutes(5),
      }),
      threshold: 1,
      evaluationPeriods: 1,
      treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
      comparisonOperator:
        cloudwatch.ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
    });
    apiErrorAlarm.addAlarmAction(alarmAction);

    // Lambda Invocation Errors Alarm for Streamer Function
    const streamerErrorAlarm = new cloudwatch.Alarm(
      this,
      "StreamerErrorAlarm",
      {
        alarmName: `${config.stackPrefix}-StreamerErrors`,
        alarmDescription: "Alert on streamer Lambda function errors",
        metric: new cloudwatch.Metric({
          namespace: "AWS/Lambda",
          metricName: "Errors",
          dimensionsMap: {
            FunctionName: props.streamerFunctionName,
          },
          statistic: "Sum",
          period: cdk.Duration.minutes(5),
        }),
        threshold: 1,
        evaluationPeriods: 1,
        treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
        comparisonOperator:
          cloudwatch.ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
      },
    );
    streamerErrorAlarm.addAlarmAction(alarmAction);

    // Structured JSON Logs Alarm with notify=true
    // Import existing log group for the API function
    // Note: Lambda automatically creates log groups on first invocation
    const apiLogGroup = logs.LogGroup.fromLogGroupName(
      this,
      "ApiLogGroup",
      `/aws/lambda/${props.apiFunctionName}`,
    );

    const apiNotifyMetric = new logs.MetricFilter(this, "ApiNotifyMetric", {
      logGroup: apiLogGroup,
      metricNamespace: `${config.stackPrefix}/Logs`,
      metricName: "ApiNotifyMessages",
      filterPattern: logs.FilterPattern.literal("{ $.notify = true }"),
      metricValue: "1",
      defaultValue: 0,
    });

    const apiNotifyAlarm = new cloudwatch.Alarm(this, "ApiNotifyAlarm", {
      alarmName: `${config.stackPrefix}-ApiNotifyMessages`,
      alarmDescription: "Alert when API logs contain notify=true",
      metric: apiNotifyMetric.metric({
        statistic: "Sum",
        period: cdk.Duration.minutes(1),
      }),
      threshold: 1,
      evaluationPeriods: 1,
      treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
      comparisonOperator:
        cloudwatch.ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
    });
    apiNotifyAlarm.addAlarmAction(alarmAction);

    // Import existing log group for the streamer function
    // Note: Lambda automatically creates log groups on first invocation
    const streamerLogGroup = logs.LogGroup.fromLogGroupName(
      this,
      "StreamerLogGroup",
      `/aws/lambda/${props.streamerFunctionName}`,
    );

    const streamerNotifyMetric = new logs.MetricFilter(
      this,
      "StreamerNotifyMetric",
      {
        logGroup: streamerLogGroup,
        metricNamespace: `${config.stackPrefix}/Logs`,
        metricName: "StreamerNotifyMessages",
        filterPattern: logs.FilterPattern.literal("{ $.notify = true }"),
        metricValue: "1",
        defaultValue: 0,
      },
    );

    const streamerNotifyAlarm = new cloudwatch.Alarm(
      this,
      "StreamerNotifyAlarm",
      {
        alarmName: `${config.stackPrefix}-StreamerNotifyMessages`,
        alarmDescription: "Alert when streamer logs contain notify=true",
        metric: streamerNotifyMetric.metric({
          statistic: "Sum",
          period: cdk.Duration.minutes(1),
        }),
        threshold: 1,
        evaluationPeriods: 1,
        treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
        comparisonOperator:
          cloudwatch.ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
      },
    );
    streamerNotifyAlarm.addAlarmAction(alarmAction);

    // API Lambda throttle alarm
    const apiThrottleAlarm = new cloudwatch.Alarm(this, "ApiThrottleAlarm", {
      alarmName: `${config.stackPrefix}-ApiThrottles`,
      alarmDescription: "Alert on API Lambda function throttles",
      metric: new cloudwatch.Metric({
        namespace: "AWS/Lambda",
        metricName: "Throttles",
        dimensionsMap: {
          FunctionName: props.apiFunctionName,
        },
        statistic: "Sum",
        period: cdk.Duration.minutes(5),
      }),
      threshold: 5,
      evaluationPeriods: 1,
      treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
      comparisonOperator: cloudwatch.ComparisonOperator.GREATER_THAN_THRESHOLD,
    });
    apiThrottleAlarm.addAlarmAction(alarmAction);

    // Streamer Lambda throttle alarm
    const streamerThrottleAlarm = new cloudwatch.Alarm(
      this,
      "StreamerThrottleAlarm",
      {
        alarmName: `${config.stackPrefix}-StreamerThrottles`,
        alarmDescription: "Alert on streamer Lambda function throttles",
        metric: new cloudwatch.Metric({
          namespace: "AWS/Lambda",
          metricName: "Throttles",
          dimensionsMap: {
            FunctionName: props.streamerFunctionName,
          },
          statistic: "Sum",
          period: cdk.Duration.minutes(5),
        }),
        threshold: 5,
        evaluationPeriods: 1,
        treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
        comparisonOperator:
          cloudwatch.ComparisonOperator.GREATER_THAN_THRESHOLD,
      },
    );
    streamerThrottleAlarm.addAlarmAction(alarmAction);

    // ReattestBatch Lambda monitoring (if provided)
    if (props.reattestBatchFunctionName) {
      // ReattestBatch Lambda error alarm
      const reattestBatchErrorAlarm = new cloudwatch.Alarm(
        this,
        "ReattestBatchErrorAlarm",
        {
          alarmName: `${config.stackPrefix}-ReattestBatchErrors`,
          alarmDescription: "Alert on reattestBatch Lambda function errors",
          metric: new cloudwatch.Metric({
            namespace: "AWS/Lambda",
            metricName: "Errors",
            dimensionsMap: {
              FunctionName: props.reattestBatchFunctionName,
            },
            statistic: "Sum",
            period: cdk.Duration.minutes(5),
          }),
          threshold: 1,
          evaluationPeriods: 1,
          treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
          comparisonOperator:
            cloudwatch.ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
        },
      );
      reattestBatchErrorAlarm.addAlarmAction(alarmAction);

      // ReattestBatch Lambda throttle alarm
      const reattestBatchThrottleAlarm = new cloudwatch.Alarm(
        this,
        "ReattestBatchThrottleAlarm",
        {
          alarmName: `${config.stackPrefix}-ReattestBatchThrottles`,
          alarmDescription: "Alert on reattestBatch Lambda function throttles",
          metric: new cloudwatch.Metric({
            namespace: "AWS/Lambda",
            metricName: "Throttles",
            dimensionsMap: {
              FunctionName: props.reattestBatchFunctionName,
            },
            statistic: "Sum",
            period: cdk.Duration.minutes(5),
          }),
          threshold: 5,
          evaluationPeriods: 1,
          treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
          comparisonOperator:
            cloudwatch.ComparisonOperator.GREATER_THAN_THRESHOLD,
        },
      );
      reattestBatchThrottleAlarm.addAlarmAction(alarmAction);

      // ReattestBatch Lambda duration alarm (alert if it takes too long)
      const reattestBatchDurationAlarm = new cloudwatch.Alarm(
        this,
        "ReattestBatchDurationAlarm",
        {
          alarmName: `${config.stackPrefix}-ReattestBatchDuration`,
          alarmDescription:
            "Alert when reattestBatch Lambda execution exceeds 10 minutes",
          metric: new cloudwatch.Metric({
            namespace: "AWS/Lambda",
            metricName: "Duration",
            dimensionsMap: {
              FunctionName: props.reattestBatchFunctionName,
            },
            statistic: "Maximum",
            period: cdk.Duration.minutes(5),
          }),
          threshold: 600000, // 10 minutes in milliseconds
          evaluationPeriods: 1,
          treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
          comparisonOperator:
            cloudwatch.ComparisonOperator.GREATER_THAN_THRESHOLD,
        },
      );
      reattestBatchDurationAlarm.addAlarmAction(alarmAction);

      // Import existing log group for the reattestBatch function
      const reattestBatchLogGroup = logs.LogGroup.fromLogGroupName(
        this,
        "ReattestBatchLogGroup",
        `/aws/lambda/${props.reattestBatchFunctionName}`,
      );

      const reattestBatchNotifyMetric = new logs.MetricFilter(
        this,
        "ReattestBatchNotifyMetric",
        {
          logGroup: reattestBatchLogGroup,
          metricNamespace: `${config.stackPrefix}/Logs`,
          metricName: "ReattestBatchNotifyMessages",
          filterPattern: logs.FilterPattern.literal("{ $.notify = true }"),
          metricValue: "1",
          defaultValue: 0,
        },
      );

      const reattestBatchNotifyAlarm = new cloudwatch.Alarm(
        this,
        "ReattestBatchNotifyAlarm",
        {
          alarmName: `${config.stackPrefix}-ReattestBatchNotifyMessages`,
          alarmDescription: "Alert when reattestBatch logs contain notify=true",
          metric: reattestBatchNotifyMetric.metric({
            statistic: "Sum",
            period: cdk.Duration.minutes(1),
          }),
          threshold: 1,
          evaluationPeriods: 1,
          treatMissingData: cloudwatch.TreatMissingData.NOT_BREACHING,
          comparisonOperator:
            cloudwatch.ComparisonOperator.GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
        },
      );
      reattestBatchNotifyAlarm.addAlarmAction(alarmAction);
    }

    // Outputs
    new cdk.CfnOutput(this, "AlertTopicArn", {
      value: this.alertTopic.topicArn,
      description: "SNS Topic ARN for alerts",
    });

    new cdk.CfnOutput(this, "AlertEmail", {
      value: config.alertEmail,
      description: "Email address for alert notifications",
    });

    // Log alarm summary
    new cdk.CfnOutput(this, "AlarmsConfigured", {
      value: JSON.stringify({
        iteratorAge: iteratorAgeAlarm.alarmName,
        apiErrors: apiErrorAlarm.alarmName,
        streamerErrors: streamerErrorAlarm.alarmName,
        apiNotify: apiNotifyAlarm.alarmName,
        streamerNotify: streamerNotifyAlarm.alarmName,
        apiThrottles: apiThrottleAlarm.alarmName,
        streamerThrottles: streamerThrottleAlarm.alarmName,
      }),
      description: "List of configured alarms",
    });
  }
}
