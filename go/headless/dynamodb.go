package headless

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func PutItemtoDynamoDB(product ProductPage) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	})
	if err != nil {
		return err
	}

	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(product)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("AmazonProducts"),
		Item:      av,
	}
	_, err = svc.PutItem(input)
	if err != nil {
		return err
	}

	fmt.Println("DynamoDBへ保存が成功しました")
	return nil
}
