package main

import (
	"context"
	"fmt"
	"image-api/internal/config"
	"image-api/internal/handler"
	"image-api/internal/service"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("설정 로드 실패: %v", err)
	}

	awsCfg, err := config.BuildAWSConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("AWS 설정 빌드 실패: %v", err)
	}

	s3Svc := service.NewS3Service(awsCfg, cfg.S3Bucket)
	ddbSvc := service.NewDynamoService(awsCfg, cfg.DynamoTable)

	imgHandler := handler.NewImageHandler(s3Svc, ddbSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("POST /images", imgHandler.Upload)
	mux.HandleFunc("GET /images/{id}", imgHandler.GetMetadata)
	mux.HandleFunc("GET /images/{id}/download", imgHandler.Download)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("이미지 API 서버 시작: %s", addr)
	log.Printf("S3 버킷: %s | DynamoDB 테이블: %s", cfg.S3Bucket, cfg.DynamoTable)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("서버 실행 실패: %s", err)
	}
}
