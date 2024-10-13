package cfr2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/yaml.v3"
	"os"
)

const ProviderName = "cloudflare-r2"

type Config struct {
	AccessKey  string `yaml:"accessKey"`
	SecretKey  string `yaml:"secretKey"`
	AccountID  string `yaml:"accountID"`
	BucketName string `yaml:"bucketName"`
}

type CloudflareR2Provider struct {
	cfg    Config
	client *s3.Client
}

func NewCloudflareR2Provider(configNode *yaml.Node) (*CloudflareR2Provider, error) {
	var provider CloudflareR2Provider

	cfg, err := decodeProviderConfigNode(configNode)
	if err != nil {
		return &provider, err
	}

	client, err := newClient(cfg)
	if err != nil {
		return &provider, err
	}

	provider.cfg = cfg
	provider.client = client
	return &provider, nil
}

func decodeProviderConfigNode(node *yaml.Node) (Config, error) {
	var cfg Config

	err := node.Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("invalid provider config: %w", err)
	}

	if cfg.AccessKey == "" {
		return cfg, fmt.Errorf("invalid provider config: missing access key")
	}
	if cfg.SecretKey == "" {
		return cfg, fmt.Errorf("invalid provider config: missing secret key")
	}
	if cfg.AccountID == "" {
		return cfg, fmt.Errorf("invalid provider config: missing account id")
	}
	if cfg.AccessKey == "" {
		return cfg, fmt.Errorf("invalid provider config: missing bucket name")
	}

	return cfg, nil
}

func newClient(cfg Config) (*s3.Client, error) {
	var client *s3.Client

	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return client, fmt.Errorf("unable to authenticate: %w", err)
	}

	client = s3.NewFromConfig(awsCfg, func(opts *s3.Options) {
		opts.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
	})
	return client, nil
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
