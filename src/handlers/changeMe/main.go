package main

import (
	"fmt"

	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type GameNames struct {
	GameName string `json:"gameName"`
}
type Games struct {
	GameName  string `json:"gameName"`
	Image     string `json: "image"`
	GameTitle string `json: "gameTitle"`
}

func main() {
	lambda.Start(Handler)
}

// Handler to A lambda will need to be created to check a user active games.
//the lambda will need to send back a json object containing a list of names of games gathered from dynamoDB
//(tournament-user-game table).
// The endpoint for the mobile to call is: https://a7sbyfy7sg.execute-api.eu-west-1.amazonaws.com/v1/user/{username}/games method = get
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	jsonRequest, _ := json.Marshal(request)
	fmt.Println(string(jsonRequest))
	fmt.Println(request.PathParameters)
	session := getDbSession()
	//rip the username from the request.
	user := request.PathParameters["username"]
	fmt.Println("User: ", user)

	gamesInProgressQuery := &dynamodb.QueryInput{
		TableName:              aws.String("tournament-user-game"),
		KeyConditionExpression: aws.String("userName = :userName"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userName": { // these are the values you will pass in to query by. must always pass in partition key
				S: aws.String(user),
			},
		},
	}
	games, _ := session.Query(gamesInProgressQuery)
	gameNames := MapGameNames(games)

	gameList := []Games{}
	for _, game := range gameNames {

		gameListQuery := &dynamodb.QueryInput{
			TableName:              aws.String("tournament-game"),
			KeyConditionExpression: aws.String("gameName = :gameName"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":gameName": { // these are the values you will pass in to query by. must always pass in partition key
					S: aws.String(game.GameName),
				},
			},
		}
		gameData, _ := session.Query(gameListQuery)
		gameList = MapGames(gameData, gameList)
	}

	fmt.Println(gameList)
	// read the games that the user is signed up to
	jsonUsers, _ := json.Marshal(gameList)
	return events.APIGatewayProxyResponse{Body: string(jsonUsers), StatusCode: 200, IsBase64Encoded: false, Headers: map[string]string{"Access-Control-Allow-Origin": "*"}}, nil
}

func MapGames(result *dynamodb.QueryOutput, games []Games) []Games {
	for _, user := range result.Items {

		gameStruct := Games{}

		err := dynamodbattribute.UnmarshalMap(user, &gameStruct)

		if err != nil {

			fmt.Println(err)

			return nil

		}

		games = append(games, gameStruct)

	}

	return games

}

func MapGameNames(result *dynamodb.QueryOutput) []GameNames {
	games := []GameNames{}
	for _, user := range result.Items {

		gameStruct := GameNames{}

		err := dynamodbattribute.UnmarshalMap(user, &gameStruct)

		if err != nil {

			fmt.Println(err)

			return nil

		}

		games = append(games, gameStruct)

	}

	return games

}

// DbSession returns a connection to dynamoDB
func getDbSession() *dynamodb.DynamoDB {
	region := "eu-west-1"
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)},
	))
	svc := dynamodb.New(sess)
	return svc
}
