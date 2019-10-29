# Cloud Function to alert notification to Slack from Stackdriver webhook notification.

## deploy

```
$ gcloud functions deploy alert_to_slack \
    --entry-point AlertToSlack \
    --runtime go111 \
    --set-env-vars 'CHANNEL=...,WEBHOOK_URL=...' \
    --trigger-http \
    --project ... \
    --region ...
```
