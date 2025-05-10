package cloud8

import (
	"context"
	"deifzar/reportingm8/pkg/configparser"
	"deifzar/reportingm8/pkg/log8"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	// "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/spf13/viper"
)

type Cloud8 struct {
	Config   *viper.Viper
	Provider string
}

func NewCloud8() (Cloud8Interface, error) {
	v, err := configparser.InitConfigParser()
	if err != nil {
		return &Cloud8{}, err
	}
	provider := v.GetString("Cloud.provider")
	return &Cloud8{Config: v, Provider: provider}, nil
}

func (m *Cloud8) UploadToBucket(file string) (string, error) {
	var err error
	var uri string
	switch m.Provider {
	case "AWS":
		uri, err = m.uploadToAWSS3Bucket(file)
	default:
		err = errors.New("not implemented yet")
	}
	return uri, err
}

func (m *Cloud8) newAWS() aws.Config {

	var config = aws.Config{
		Region:      *aws.String(m.Config.GetString("Cloud.AWS.region")),
		Credentials: credentials.NewStaticCredentialsProvider(m.Config.GetString("Cloud.AWS.key"), m.Config.GetString("Cloud.AWS.secret"), ""),
	}

	return config

	// sess, err := session.NewSession(&aws.Config{
	// 	Region:      aws.String("us-west-2"),
	// 	Credentials: credentials.NewStaticCredentials(m.Config.GetString("Cloud.AWS.key"), m.Config.GetString("Cloud.AWS.secret"), ""),
	// })
	// cfg, err := config.LoadDefaultConfig(context.Background(),
	// config.WithRegion(m.Config.GetString("Cloud.AWS.region")),
	// config.WithSharedConfigProfile(profileName))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// return cfg
	// }
}

func (m *Cloud8) uploadToAWSS3Bucket(fileName string) (string, error) {

	file, err := os.Open(fileName)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		log8.BaseLogger.Error().Msg("errors triggered when reading the file before uploading it into the s3 bucket")
		return "", err
	}
	defer file.Close()
	bucketName := m.Config.GetString("Cloud.AWS.bucket")
	objectKey := filepath.Base(fileName)
	ctx := context.Background()
	cfg := m.newAWS()
	s3client := s3.NewFromConfig(cfg)

	_, err = s3client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			log8.BaseLogger.Error().Msgf("Error while uploading object to %s. The object is too large.\n"+
				"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
				"or the multipart upload API (5TB max).", bucketName)
		} else {
			log8.BaseLogger.Error().Msgf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				fileName, bucketName, objectKey, err)
		}
		return "", err
	}
	err = s3.NewObjectExistsWaiter(s3client).Wait(
		ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(objectKey)}, time.Minute)
	if err != nil {
		log8.BaseLogger.Error().Msgf("Failed attempt to wait for object %s to exist.\n", objectKey)
	}
	// s3://report-cptm8/wood.png
	var s3uri = "s3://" + m.Config.GetString("Cloud.AWS.region") + "/" + bucketName + "/" + objectKey
	return s3uri, err
}

// presignClient := s3.NewPresignClient(s3client)
// presignedUrl, err := presignClient.PresignPutObject(context.Background(),
// 	&s3.PutObjectInput{
// 	Bucket: aws.String(bucketName),
// 	Key:    aws.String(objectName),
// 	},
// 	s3.WithPresignExpires(time.Minute*15))
// 	if err != nil {
// 	log.Fatal(err)
// 	}
// return presignedUrl.URL
