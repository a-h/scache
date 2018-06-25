package dynamo

import (
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/a-h/scache/example/user"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// NewUserStore creates a new UserStore with required fields populated.
func NewUserStore(region, tableName string) (store *UserStore, err error) {
	conf := &aws.Config{
		Region: aws.String(region),
	}
	sess, err := session.NewSession(conf)
	if err != nil {
		return
	}
	store = &UserStore{
		Client:    dynamodb.New(sess),
		TableName: tableName,
	}
	return
}

const (
	// UserTableName is the name of the 'name' column in DynamoDB.
	UserTableName = "name"
	// UserTableEmail is the name of the 'email' column in DynamoDB.
	UserTableEmail = "email"
)

// UserTableKey is the definition of the hash key for the User table.
func UserTableKey(id string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"id": {
			S: aws.String(id),
		},
	}
}

// UserStore handles user data in DynamoDB.
type UserStore struct {
	Client    *dynamodb.DynamoDB
	TableName string
}

// Put upserts data into DynamoDB.
func (us UserStore) Put(u user.User) (err error) {
	item, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return
	}
	_, err = us.Client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(us.TableName),
		Item:      item,
	})
	return
}

// Get retrieves data from DynamoDB.
func (us UserStore) Get(id string) (u user.User, ok bool, err error) {
	gii := &dynamodb.GetItemInput{
		TableName: aws.String(us.TableName),
		Key:       UserTableKey(id),
	}

	out, err := us.Client.GetItem(gii)
	if err != nil {
		return
	}

	err = dynamodbattribute.UnmarshalMap(out.Item, &u)
	ok = u.ID != ""
	return
}

// Delete deletes a user's record.
func (us UserStore) Delete(id string) (err error) {
	d := &dynamodb.DeleteItemInput{
		TableName: aws.String(us.TableName),
		Key:       UserTableKey(id),
	}
	_, err = us.Client.DeleteItem(d)
	return
}
