#!/bin/bash
# DynamoDB 테이블 생성 스크립트
# 사용법: ./scripts/create-table.sh
#
# 사전 조건: AWS CLI 설정 완료 (aws configure)

set -e

TABLE_NAME="${DYNAMODB_TABLE_NAME:-images}"
REGION="${AWS_REGION:-ap-northeast-2}"

echo "DynamoDB 테이블 생성 중: $TABLE_NAME (리전: $REGION)"

aws dynamodb create-table \
  --table-name "$TABLE_NAME" \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region "$REGION"

echo "테이블 활성화 대기 중..."
aws dynamodb wait table-exists --table-name "$TABLE_NAME" --region "$REGION"

echo "완료! 테이블 '$TABLE_NAME' 이 생성되었습니다."
