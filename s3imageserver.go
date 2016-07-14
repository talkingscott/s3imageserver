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

	"github.com/nfnt/resize"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	region := os.Getenv("IMAGEBUCKETREGION")
	bucket := os.Getenv("IMAGEBUCKET")
	fmt.Fprintf(os.Stdout, "Region: %s bucket: %s\n", region, bucket)

	svc := s3.New(session.New(&aws.Config{Region: aws.String(region)}))

	http.HandleFunc("/2/", func(w http.ResponseWriter, r *http.Request) {
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

		// get image from bucket
		// TODO: if If-Modified-Since header is present, compare to object date and return 304 if unmodified
		params := &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key: aws.String(pieces[3]),
		}
		resp, err := svc.GetObject(params)
		
		if err != nil {
			http.Error(w, err.Error(), 404)
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
		// TODO: set last-modified to file modified date

    	// encode image and stream as response
		jpeg.Encode(w, m, nil)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
