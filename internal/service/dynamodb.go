package service

import (
	"context"
	"fmt"
	"image-server/internal/model"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoService struct {
	client *dynamodb.Client
	table  string
}

func NewDynamoService(cfg aws.Config, table string) *DynamoService {
	return &DynamoService{
		client: dynamodb.NewFromConfig(cfg),
		table:  table,
	}
}

func (d *DynamoService) PutImage(ctx context.Context, img *model.Image) error {
	log.Printf("[DynamoDB] PutItem: table=%s, id=%s", d.table, img.ID)

	item, err := attributevalue.MarshalMap(img)
	if err != nil {
		return fmt.Errorf("DynamoDB 마샬링 실패: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.table),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("DynamoDB PutItem 실패: %w", err)
	}

	log.Printf("[DynamoDB] PutItem 완료: id=%s", img.ID)
	return nil
}

func (d *DynamoService) GetImage(ctx context.Context, id string) (*model.Image, error) {
	log.Printf("[DynamoDB] GetItem: table=%s, id=%s", d.table, id)

	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.table),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("DynamoDB GetItem 실패: %w", err)
	}

	if result.Item == nil {
		log.Printf("[DynamoDB] GetItem: 항목 없음 (id=%s)", id)
		return nil, nil
	}

	var img model.Image
	if err := attributevalue.UnmarshalMap(result.Item, &img); err != nil {
		return nil, fmt.Errorf("DynamoDB 언마샬링 실패: %w", err)
	}

	log.Printf("[DynamoDB] GetItem 완료: id=%s, file=%s", img.ID, img.FileName)
	return &img, nil
}
