# s3imageserver
A simple web server for images in S3 that supports resizing (scaling)

## Invoking

The URL pattern is /2/image/:image/width/:width

## Configuration

Application configuration is through the environment.

```
export IMAGEBUCKET=image-bucket-name
export IMAGEBUCKETREGION=us-east-1
```

AWS configuration is AWS SDK for go standard, namely ~/.aws/credentials for credentials.
