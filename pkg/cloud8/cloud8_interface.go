package cloud8

import "github.com/aws/aws-sdk-go-v2/aws"

type Cloud8Interface interface {
	UploadToBucket(file string) (string, error)
	// return s3 URI
	uploadToAWSS3Bucket(file string) (string, error)
	newAWS() aws.Config
}
