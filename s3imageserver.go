package main

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	bucket := os.Getenv("IMAGEBUCKET")
	fmt.Fprintf(os.Stdout, "Bucket: %s\n", bucket)

	// let SDK set region via AWS_REGION
	//config := aws.Config{}
	//config.WithLogLevel(aws.LogDebugWithHTTPBody)
	//svc := s3.New(session.New(&config))
	svc := s3.New(session.New(&aws.Config{}))

	http.HandleFunc("/2/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: support HEAD
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		// parse URL
		pieces := strings.Split(r.URL.Path, "/")
		if len(pieces) != 6 || pieces[2] != "image" || pieces[4] != "width" {
			http.NotFound(w, r)
			return
		}

		width, err := strconv.ParseUint(pieces[5], 10, 32);
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		ifModifiedSince := r.Header.Get("if-modified-since")
		fmt.Fprintf(os.Stdout, "if-modified-since: %s\n", ifModifiedSince)

		// get image from bucket
		params := s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key: aws.String(pieces[3]),
		}
		if ifModifiedSince != "" {
			tm, err := time.Parse(time.RFC1123, ifModifiedSince)
			if err != nil {
				fmt.Fprintf(os.Stdout, "parse error %s\n", err)
			} else {
				params.IfModifiedSince = aws.Time(tm)
				fmt.Fprintf(os.Stdout, "set IfModifiedSince to %s\n", params.IfModifiedSince.Format(time.RFC1123))
			}
		}
		resp, err := svc.GetObject(&params)
		
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// Get error details
				fmt.Fprintf(os.Stdout, "AWS Code: %s Message: %s\n", awsErr.Code(), awsErr.Message())

				// Prints out full error message, including original error if there was one.
				fmt.Fprintf(os.Stdout, "AWS Error: %s\n", awsErr.Error())
		 
				// Get original error
				if origErr := awsErr.OrigErr(); origErr != nil {
					// operate on original error.
				}
				if awsErr.Code() == "304NotModified" {
					http.Error(w, awsErr.Message(), 304)
				} else {
					http.Error(w, awsErr.Message(), 404)
				}
			} else {
				fmt.Fprintf(os.Stdout, "Plain error: %s\n", err.Error())
				http.Error(w, err.Error(), 404)
			}
			return
		}
		
		// decode image
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// resize to width using Lanczos resampling and preserve aspect ratio
		m := resize.Resize(uint(width), 0, img, resize.Lanczos3)

		// set headers
		w.Header().Set("content-type", "image/jpeg")
		w.Header().Set("cache-control", "max-age=86400")
		w.Header().Set("last-modified", resp.LastModified.Format(time.RFC1123))

		// encode image and stream as response
		jpeg.Encode(w, m, nil)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
