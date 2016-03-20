package main

import (
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	svc := s3.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))
		
	http.HandleFunc("/2/", func(w http.ResponseWriter, r *http.Request) {
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

		// if If-Modified-Since header is present, compare to file date and return 304 if unmodified
		params := &s3.GetObjectInput{
			Bucket:                     aws.String("com.scottnichol.s3test"), // Required
			Key:                        aws.String(pieces[3]),  // Required
		}
		resp, err := svc.GetObject(params)
		
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		
		// decode jpeg into image.Image
		img, err := jpeg.Decode(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// resize to width using Lanczos resampling
		// and preserve aspect ratio
		m := resize.Resize(uint(width), 0, img, resize.Lanczos3)

		// set headers
		w.Header().Set("content-type", "image/jpeg")
		w.Header().Set("cache-control", "max-age=86400")
		// TODO: set last-modified to file modified date

    		// write new image 
		jpeg.Encode(w, m, nil)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

