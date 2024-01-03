package upload

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

type UploadHandlerResponse struct {
	id     string  `json:"url"`
	Error  *string `json:"error,omitempty"` // Optional field
	Status Status  `json:"status"`          // "error" or "success"
}

func UploadToStorage(file multipart.File, sign string, contentType string) (UploadHandlerResponse, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", os.Getenv("STORAGE_ACCOUNT_ID")),
		}, nil
	})

	// AWS SDK Configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("STORAGE_ACCESS_KEY"), os.Getenv("STORAGE_SECRET"), "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
		errorReponse := UploadHandlerResponse{
			id:     "",
			Status: StatusSuccess,
			Error:  aws.String(err.Error()),
		}
		return errorReponse, err
	}

	// Create an S3 client
	client := s3.NewFromConfig(cfg)
	id := uuid.New()
	filename := fmt.Sprintf("sign_%s_raw", id.String())

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
		errorReponse := UploadHandlerResponse{
			id:     "",
			Status: StatusSuccess,
			Error:  aws.String(err.Error()),
		}
		return errorReponse, err
	}

	// Construct the URL of the uploaded file
	successResponse := UploadHandlerResponse{
		id:     id.String(),
		Status: StatusSuccess,
	}

	return successResponse, nil
}
