package s3

import (
	"bytes"
	"cms-utils-go/connector/aws/cred"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// ============ Constants =============

// S3 ACL constants
const (
	PublicRead = "public-read"
	Private    = "private"
)

const GenerateDirectoryPermissionMode = 0750
const FileWritePermissionMode = 0644
const FileReadPermissionMode = 0644

type S3Client struct {
	cred.Cred
	session    *session.Session
	s3         *s3.S3
	BucketName string
}

// =========== Exposed (public) Methods - can be called from external packages ============

// UploadFile uploads the input file (inputFilePath) into S3.
// it stores the file at "s3Location" location with "s3FileName" file name
// with "acl" permission
// acl - "private", "public-read", "etc"
func (s3Client *S3Client) UploadFile(inputFilePath, s3Location, s3FileName, acl string) (location string, err error) {
	file, err := os.OpenFile(inputFilePath, os.O_RDONLY, os.FileMode(GenerateDirectoryPermissionMode))
	if err != nil {
		return
	}
	defer file.Close()
	// creates a new s3 session
	readFile, err := ioutil.ReadAll(file)
	fileBytes := bytes.NewReader(readFile)
	fileType := http.DetectContentType(readFile)

	// uploads the file
	s3PathKey := s3Location + s3FileName
	s3Uploader := s3manager.NewUploader(s3Client.session)
	s3UploadInput := &s3manager.UploadInput{
		Bucket:      aws.String(s3Client.BucketName),
		Key:         aws.String(s3PathKey),
		Body:        fileBytes,
		ACL:         aws.String(acl),
		ContentType: aws.String(fileType),
	}
	result, err := s3Uploader.Upload(s3UploadInput)
	if err != nil {
		return
	}
	location = result.Location
	return
}

// DownloadFile gets the file content from the S3 location (s3key)
// and stores them in the mentioned "outputFilePath".
func (s3Client *S3Client) DownloadFile(outputFilePath, s3key string) (err error) {
	file, err := os.Create(outputFilePath)
	if err != nil {
		return
	}
	defer file.Close()

	// downloads the file
	s3Downloader := s3manager.NewDownloader(s3Client.session)
	_, err = s3Downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s3Client.BucketName),
			Key:    aws.String(s3key),
		})
	if err != nil {
		return
	}
	return
}

// DeleteFile deletes the files from the S3 location (s3key)
func (s3Client *S3Client) DeleteFiles(s3keys []string) (err error) {
	for _, s3Key := range s3keys {
		if _, err = s3Client.s3.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s3Client.BucketName),
			Key:    aws.String(s3Key),
		}); err != nil {
			return
		}
	}

	return
}

// RemoveFile deletes a single file from the S3 location
func (s3Client *S3Client) RemoveFile(key string) (err error) {
	if key == "" {
		return
	}

	deleteS3Object := s3.DeleteObjectInput{
		Bucket: aws.String(s3Client.BucketName),
		Key:    aws.String(key),
	}
	_, err = s3Client.s3.DeleteObject(&deleteS3Object)
	if err != nil {
		return
	}
	return
}

// GetPreSignFile generates temp url for the file for the specified duration
func (s3Client *S3Client) GetPreSignFile(filePath string, duration time.Duration) (urlStr string, err error) {

	// Create S3 service client

	req, _ := s3Client.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s3Client.BucketName),
		Key:    aws.String(filePath),
	})
	urlStr, err = req.Presign(duration)

	if err != nil {
		return
	}
	return
}

// New create a s3 client to interact with the specified Bucket on S3
func (s3Client *S3Client) New() (err error) {

	var s3Config aws.Config
	s3Config.Region = aws.String(s3Client.Region)
	if s3Client.Key != "" && s3Client.Secret != "" {
		s3Config.Credentials = credentials.NewStaticCredentials(s3Client.Key, s3Client.Secret, "")
	}

	s3Client.session, err = session.NewSession(&s3Config)
	if err != nil {
		return
	}

	s3Client.s3 = s3.New(s3Client.session)
	return
}

// ============ Internal(private) Methods - can only be called from inside this package ==============
