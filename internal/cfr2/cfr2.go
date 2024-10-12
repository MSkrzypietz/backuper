package cfr2

import (
	"context"
	"fmt"
	"github.com/MSkrzypietz/backuper/internal/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"
)

const ProviderName = "cloudflare-r2"

type CloudflareR2Provider struct {
	cfg    storage.ProviderConfig
	client *s3.Client
}

func NewCloudflareR2Provider(cfg storage.ProviderConfig) (*CloudflareR2Provider, error) {
	var provider CloudflareR2Provider
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return &provider, fmt.Errorf("unable to authenticate: %w", err)
	}

	provider.cfg = cfg
	provider.client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
	})
	return &provider, nil
}

func (p *CloudflareR2Provider) GetName() string {
	return ProviderName
}

func (p *CloudflareR2Provider) Backup(name string, zipPath string) error {
	return p.uploadToR2(name, zipPath)
}

func (p *CloudflareR2Provider) uploadToR2(key string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("unable to get file info: %v", err)
	}

	_, err = p.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(p.cfg.BucketName),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
	})

	if err != nil {
		return fmt.Errorf("unable to upload %q to bucket %q, %v", filePath, p.cfg.BucketName, err)
	}

	return nil
}
