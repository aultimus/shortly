package db

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	tableName = "URL"
)

// DynamoService satisfies the DBer interface
type DynamoService struct {
	svc *dynamodb.DynamoDB
}

func NewDynamoService() *DynamoService {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)
	if err != nil {
		panic(err.Error())
	}
	svc := dynamodb.New(sess)
	return &DynamoService{svc}
}

type Item struct {
	// slightly annoying that capitalisation is inconsistent amongst keys but no biggy!
	Hash        string `json:"Hash"`
	OriginalURL string `json:"original_url"`
}

func (d *DynamoService) Create(key string, data *StoredURL) error {
	item := Item{key, data.OriginalURL}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return NewErrDB(err.Error())
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = d.svc.PutItem(input)
	if err != nil {
		err = NewErrDB(err.Error())
	}
	return err
}

func (d *DynamoService) Get(key string) (*StoredURL, error) {
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Hash": {
				S: aws.String(key),
			},
		},
	})

	if err != nil {
		return nil, NewErrDB(err.Error())
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, NewErrDB(err.Error())
	}

	if item.Hash == "" {
		return nil, fmt.Errorf("could not find key %s", key)
	}

	return &StoredURL{item.OriginalURL}, nil
}
