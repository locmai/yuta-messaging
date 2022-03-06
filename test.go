package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

var homeserver = flag.String("homeserver", "", "Matrix homeserver")
var username = flag.String("username", "", "Matrix username localpart")

var password = flag.String("password", "", "Matrix password")

func maintest() {
	flag.Parse()
	if *username == "" || *password == "" || *homeserver == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Logging into", *homeserver, "as", *username)
	client, err := mautrix.NewClient(*homeserver, "", "")
	if err != nil {
		panic(err)
	}
	_, err = client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: *username},
		Password:         *password,
		StoreCredentials: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Login successful")

	start := time.Now().UnixMilli()
	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		if evt.Sender == "@locmai:dendrite.maibaloc.com" && evt.Timestamp > start {
			fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body)

			fulfillmentText, err := DetectIntentText("yuta-seig", "test", evt.Content.AsMessage().Body)
			if err != nil {
				panic(err)
			}
			client.SendText(evt.RoomID, fulfillmentText)
		}
	})

	err = client.Sync()
	if err != nil {
		panic(err)
	}
}

func DetectIntentTextTest(projectID, sessionID, text string) (string, error) {
	ctx := context.Background()
	languageCode := "en-US"

	sessionClient, err := dialogflow.NewSessionsClient(ctx)
	if err != nil {
		return "", err
	}
	defer sessionClient.Close()

	if projectID == "" || sessionID == "" {
		return "", fmt.Errorf("received empty project (%s) or session (%s)", projectID, sessionID)
	}

	sessionPath := fmt.Sprintf("projects/%s/agent/sessions/%s", projectID, sessionID)
	textInput := dialogflowpb.TextInput{Text: text, LanguageCode: languageCode}
	queryTextInput := dialogflowpb.QueryInput_Text{Text: &textInput}
	queryInput := dialogflowpb.QueryInput{Input: &queryTextInput}
	request := dialogflowpb.DetectIntentRequest{Session: sessionPath, QueryInput: &queryInput}

	response, err := sessionClient.DetectIntent(ctx, &request)
	if err != nil {
		return "", err
	}

	queryResult := response.GetQueryResult()
	fulfillmentText := queryResult.GetFulfillmentText()
	return fulfillmentText, nil
}
