package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const functionDir = "../transcribe-function"

type LambdaTranscribeAudioToTextGolangStackProps struct {
	awscdk.StackProps
}

func NewLambdaTranscribeAudioToTextGolangStack(scope constructs.Construct, id string, props *LambdaTranscribeAudioToTextGolangStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	sourceBucket := awss3.NewBucket(stack, jsii.String("audio-file-source-bucket"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		//BucketName:        jsii.String("audio-file-source-bucket"),
	})

	outputBucket := awss3.NewBucket(stack, jsii.String("transcribe-job-output-bucket"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		//BucketName:        jsii.String("transcribe-job-output-bucket"),
	})

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("audio-to-text-function"),
		&awscdklambdagoalpha.GoFunctionProps{
			Runtime:     awslambda.Runtime_GO_1_X(),
			Environment: &map[string]*string{"OUTPUT_BUCKET_NAME": outputBucket.BucketName()},
			Entry:       jsii.String(functionDir),
		})

	//the roles that are granted are more than what's required. homework for reader to make this fine-grained

	sourceBucket.GrantRead(function, "*")
	outputBucket.GrantReadWrite(function, "*")
	function.Role().AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonTranscribeFullAccess")))

	function.AddEventSource(awslambdaeventsources.NewS3EventSource(sourceBucket, &awslambdaeventsources.S3EventSourceProps{
		Events: &[]awss3.EventType{awss3.EventType_OBJECT_CREATED},
	}))

	awscdk.NewCfnOutput(stack, jsii.String("audio-file-source-bucket-name"),
		&awscdk.CfnOutputProps{
			ExportName: jsii.String("audio-file-source-bucket-name"),
			Value:      sourceBucket.BucketName()})

	awscdk.NewCfnOutput(stack, jsii.String("transcribe-job-bucket-name"),
		&awscdk.CfnOutputProps{
			ExportName: jsii.String("transcribe-job-bucket-name"),
			Value:      outputBucket.BucketName()})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewLambdaTranscribeAudioToTextGolangStack(app, "LambdaTranscribeAudioToTextGolangStack", &LambdaTranscribeAudioToTextGolangStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
