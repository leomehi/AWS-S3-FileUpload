package main

import (
	"bytes"
	"context"

	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// BucketBasics encapsulates the Amazon Simple Storage Service (Amazon S3) actions
type BucketBasics struct {
	S3Client *s3.Client
}

// CreateBucket creates a bucket with the specified name in the specified Region.
func (basics BucketBasics) CreateBucket(name string, region string) error {
	_, err := basics.S3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	})
	if err != nil {
		log.Printf("Couldn't create bucket %v in Region %v. Here's why: %v\n", name, region, err)
	}
	return err
}

// UploadFileToS3 uploads a file to an S3 bucket
func (basics BucketBasics) UploadFileToS3(bucketName string, fileName string, fileData []byte) error {
	_, err := basics.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		log.Printf("Couldn't upload file to %v:%v. Here's why: %v\n", bucketName, fileName, err)
	}
	return err
}

// Handler is the main Lambda function handler
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Generate a unique bucket name based on the current timestamp
	bucketName := "your-prefix-" + time.Now().Format("20060102-150405")

	// Create S3 bucket
	basics := BucketBasics{S3Client: s3Client}
	err = basics.CreateBucket(bucketName, "ap-south-1")
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Generate a unique file name based on the current timestamp
	fileName := "upload-" + time.Now().Format("20060102-150405")

	// Upload file to S3 bucket
	err = basics.UploadFileToS3(bucketName, fileName, []byte(request.Body))
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Return a success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "File successfully uploaded to S3.",
	}, nil
}

func main() {
	lambda.Start(Handler)
}
