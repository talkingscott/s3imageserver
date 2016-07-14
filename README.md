# s3imageserver
A simple web server for images in S3 that supports resizing (scaling)

## Invoking

The URL pattern is /2/image/:image/width/:width

## Configuration

AWS configuration is AWS SDK for go standard, namely ~/.aws/credentials for credentials
and environment variable AWS_REGION for region.

Application configuration is through the environment.  So you probably would have something like

```
export IMAGEBUCKET=image-bucket-name
export AWS_REGION=us-east-1
```

