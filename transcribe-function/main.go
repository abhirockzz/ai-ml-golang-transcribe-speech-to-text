package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
)

var outputBucket string

// var s3Client *s3.Client
var transcribeClient *transcribe.Client

func init() {
	outputBucket = os.Getenv("OUTPUT_BUCKET_NAME")

	if outputBucket == "" {
		log.Fatal("missing environment variable OUTPUT_BUCKET_NAME")
	}
	fmt.Println("output S3 bucket", outputBucket)

	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		log.Fatal("failed to load config ", err)
	}

	//s3Client = s3.NewFromConfig(cfg)
	transcribeClient = transcribe.NewFromConfig(cfg)

}

func handler(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {

		fmt.Println("file", record.S3.Object.Key, "uploaded to", record.S3.Bucket.Name)

		sourceBucketName := record.S3.Bucket.Name
		fileName := record.S3.Object.Key

		err := audioToText(sourceBucketName, fileName)

		if err != nil {
			log.Fatal("failed to process file ", record.S3.Object.Key, " in bucket ", record.S3.Bucket.Name, err)
		}
	}
}

func main() {
	lambda.Start(handler)
}

func audioToText(sourceBucketName, fileName string) error {

	inputFileNameFormat := "s3://%s/%s"
	inputFile := fmt.Sprintf(inputFileNameFormat, sourceBucketName, fileName)

	languageCode := "en-US"
	jobName := "job-" + sourceBucketName + "-" + fileName

	outputFileName := strings.Split(fileName, ".")[0] + "-job-output.txt"

	_, err := transcribeClient.StartTranscriptionJob(context.Background(), &transcribe.StartTranscriptionJobInput{
		TranscriptionJobName: &jobName,
		LanguageCode:         types.LanguageCode(languageCode),
		MediaFormat:          types.MediaFormatMp3,
		Media: &types.Media{
			MediaFileUri: &inputFile,
		},
		OutputBucketName: aws.String(outputBucket),
		OutputKey:        aws.String(outputFileName),
		Settings: &types.Settings{
			ShowSpeakerLabels: aws.Bool(true),
			MaxSpeakerLabels:  aws.Int32(5),
		},
	})

	if err != nil {
		return err
	}

	fmt.Println("submitted transcribe job for file", fileName, "in bucket", sourceBucketName)

	return nil
}
