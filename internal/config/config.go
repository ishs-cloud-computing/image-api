package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AppConfig struct {
	AWS         AWSConfig
	S3Bucket    string
	DynamoTable string
	Port        string
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type awsSecret struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
}

func Load(ctx context.Context) (*AppConfig, error) {
	region := getEnvOrDefault("AWS_REGION", "ap-northeast-2")
	bucket := mustGetEnv("S3_BUCKET_NAME")
	table := mustGetEnv("DYNAMO_TABLE_NAME")
	port := getEnvOrDefault("PORT", "8080")

	log.Printf("[Config] 설정 로드 시작: region=%s, bucket=%s, table=%s, port=%s",
		region, bucket, table, port)

	var awsCfg AWSConfig
	awsCfg.Region = region

	secretName := os.Getenv("AWS_SECRET_NAME")
	if secretName != "" {
		log.Printf("[Config] Secrets Manager에서 자격증명 로드: secret=%s", secretName)
		cred, err := loadFromSecretManager(ctx, region, secretName)
		if err != nil {
			return nil, fmt.Errorf("secrets manager 로드 실패: %w", err)
		}
		awsCfg.AccessKeyID = cred.AccessKeyID
		awsCfg.SecretAccessKey = cred.SecretAccessKey
		log.Printf("[Config] Secrets Manager 로드 완료")
	} else {
		log.Printf("[Config] 환경변수에서 자격증명 로드")
		awsCfg.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
		awsCfg.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	log.Printf("[Config] 설정 로드 완료")
	return &AppConfig{
		AWS:         awsCfg,
		S3Bucket:    bucket,
		DynamoTable: table,
		Port:        port,
	}, nil
}

func BuildAWSConfig(ctx context.Context, cfg *AppConfig) (aws.Config, error) {
	if cfg.AWS.AccessKeyID != "" && cfg.AWS.SecretAccessKey != "" {
		return config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.AWS.Region),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					cfg.AWS.AccessKeyID,
					cfg.AWS.SecretAccessKey,
					"",
				),
			),
		)
	}

	return config.LoadDefaultConfig(ctx, config.WithRegion(cfg.AWS.Region))
}

func loadFromSecretManager(ctx context.Context, region, secretName string) (*awsSecret, error) {
	log.Printf("[Config] Secrets Manager API 호출: region=%s, secret=%s", region, secretName)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		log.Printf("[Config] Secrets Manager 호출 실패: %v", err)
		return nil, err
	}

	var secret awsSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return nil, fmt.Errorf("시크릿 JSON 파싱 실패: %w", err)
	}

	log.Printf("[Config] Secrets Manager에서 자격증명 파싱 완료")
	return &secret, nil

}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("필수 환경변수 %s 가 설정되지 않았습니다.", key))
	}

	return v
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultValue
}
