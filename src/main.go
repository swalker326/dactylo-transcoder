package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// func printMultipartFormData(r *http.Request) {
// 	// Specify a maximum upload size
// 	const maxUploadSize = 10 << 20 // e.g., 10 MB
// 	r.ParseMultipartForm(maxUploadSize)

// 	// Print form values
// 	fmt.Println("Form Values:")
// 	if r.MultipartForm != nil {
// 		for key, values := range r.MultipartForm.Value {
// 			for _, value := range values {
// 				fmt.Println(key, ":", value)
// 			}
// 		}
// 	}

// 	// Print file information
// 	fmt.Println("Form Files:")
// 	if r.MultipartForm != nil {
// 		for key, files := range r.MultipartForm.File {
// 			for _, file := range files {
// 				fmt.Println(key, ":", file.Filename, file.Size, "bytes")
// 			}
// 		}
// 	}
// }

// func printFormData(r *http.Request) {
// 	// Parse the form data
// 	err := r.ParseForm()
// 	if err != nil {
// 		fmt.Println("Error parsing form:", err)
// 		return
// 	}

// 	// Print form values
// 	fmt.Println("Form Data:")
// 	for key, values := range r.Form { // range over map
// 		for _, value := range values { // range over []string
// 			fmt.Println(key, ":", value)
// 		}
// 	}
// }

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	// printMultipartFormData(r)
	// Parse Multipart Form
	// log.Fatal("Parsing multipart form data")
	r.ParseMultipartForm(10 << 20) // 10 MB max
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	sign := r.FormValue("sign")
	defer file.Close()

	// You can get the content type from the FileHeader
	contentType := handler.Header.Get("Content-Type")

	// Upload to R2
	url, err := uploadToR2(file, sign, contentType)
	if err != nil {
		fmt.Fprintf(w, "Failed to upload: %s\n", err)
		return
	}

	// Return the URL of the uploaded file
	fmt.Fprintf(w, "File uploaded successfully: %s\n", url)
}

func uploadToR2(file multipart.File, sign string, contentType string) (string, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", os.Getenv("STORAGE_ACCOUNT_ID")),
		}, nil
	})

	// AWS SDK Configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("STORAGE_ACCECC_KEY"), os.Getenv("STORAGE_SECRET"), "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)
	filename := fmt.Sprintf("sign_%s", uuid.New())

	// Create the PutObjectInput
	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("STORAGE_BUCKET")),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
	}

	// Upload the file
	_, err = client.PutObject(context.TODO(), putObjectInput)
	if err != nil {
		return "", err
	}

	// Construct the URL of the uploaded file
	url := fmt.Sprintf("https://media.dactylo.io/%s", filepath.Base(filename))

	return url, nil
}
