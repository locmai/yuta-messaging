package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"github.com/gorilla/mux"
	"github.com/locmai/yuta-messaging/clients"
	"github.com/locmai/yuta-messaging/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func main() {
	cfg := config.ParseFlags()

	router := mux.NewRouter()

	router.Path("/metrics").Handler(promhttp.Handler())
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		WriteTimeout: time.Duration(cfg.Server.Timeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
	}
	serverStartTime := time.Now().UnixMilli()

	for _, clientConfig := range cfg.Clients {
		switch clientType := clientConfig.ClientType; clientType {
		case config.MatrixType:
			botClient, err := clients.NewMatrixClient(clientConfig)
			if err != nil {
				panic(err)
			}
			syncer := botClient.Client.Syncer.(*mautrix.DefaultSyncer)
			// TODO: Implement OnEventType handler
			syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
				if evt.Sender == "@locmai:dendrite.maibaloc.com" && evt.Timestamp > serverStartTime {
					fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body)

					_, action, err := DetectIntentText("yuta-seig", "test", evt.Content.AsMessage().Body)
					if err != nil {
						panic(err)
					}
					botClient.Client.SendText(evt.RoomID, action)
				}
			})
			go botClient.Client.Sync()
		case config.SlackType:
			fmt.Printf("Client is not implemented %s", clientConfig.ClientType)
		default:
			fmt.Printf("Client is not implemented %s", clientConfig.ClientType)
		}
	}

	logrus.Println("Server started")
	logrus.Fatal(srv.ListenAndServe())
}

func DetectIntentText(projectID, sessionID, text string) (string, string, error) {
	ctx := context.Background()
	languageCode := "en-US"

	sessionClient, err := dialogflow.NewSessionsClient(ctx)
	if err != nil {
		return "", "", err
	}
	defer sessionClient.Close()

	if projectID == "" || sessionID == "" {
		return "", "", fmt.Errorf("received empty project (%s) or session (%s)", projectID, sessionID)
	}

	sessionPath := fmt.Sprintf("projects/%s/agent/sessions/%s", projectID, sessionID)
	textInput := dialogflowpb.TextInput{Text: text, LanguageCode: languageCode}
	queryTextInput := dialogflowpb.QueryInput_Text{Text: &textInput}
	queryInput := dialogflowpb.QueryInput{Input: &queryTextInput}
	request := dialogflowpb.DetectIntentRequest{Session: sessionPath, QueryInput: &queryInput}

	response, err := sessionClient.DetectIntent(ctx, &request)
	if err != nil {
		return "", "", err
	}

	queryResult := response.GetQueryResult()
	fulfillmentText := queryResult.GetFulfillmentText()
	actionDetected := queryResult.Action
	return fulfillmentText, actionDetected, nil
}
