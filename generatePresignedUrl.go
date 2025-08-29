package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	PreSClient := s3.NewPresignClient(s3Client)
	params := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	presignedRequest, err := PreSClient.PresignGetObject(context.Background(), &params, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}
	return presignedRequest.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	BucketAndKey := strings.Split(*video.VideoURL, ",")
	if len(BucketAndKey) < 2 {
		return video, nil
	}
	Bucket := BucketAndKey[0]
	Key := BucketAndKey[1]

	duration := time.Duration(5 * time.Minute)

	PSURLUpdate, err := generatePresignedURL(cfg.s3Client, Bucket, Key, duration)
	if err != nil {
		return database.Video{}, fmt.Errorf("unable to generate presigned URL")
	}

	video.VideoURL = &PSURLUpdate

	return video, nil
}
