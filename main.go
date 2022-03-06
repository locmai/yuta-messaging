package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"github.com/gorilla/mux"
	"github.com/locmai/yuta-messaging/clients"
	"github.com/locmai/yuta-messaging/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
	"gopkg.in/yaml.v2"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

// Load a yaml config file for the server
// Checks the config to ensure that it is valid.
func LoadConfig(configPath string) (*config.ConfigFile, error) {
	var configFile config.ConfigFile

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	// Pass the current working directory and ioutil.ReadFile so that they can
	// be mocked in the tests
	if err = yaml.Unmarshal(configData, &configFile); err != nil {
		return nil, err
	}
	return &configFile, nil
}

func main() {
	cfg := config.ParseFlags()

	router := mux.NewRouter()

	router.Path("/metrics").Handler(promhttp.Handler())
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
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
			syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
				if evt.Sender == "@locmai:dendrite.maibaloc.com" && evt.Timestamp > serverStartTime {
					fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type.String(), evt.ID, evt.Content.AsMessage().Body)

					fulfillmentText, err := DetectIntentText("yuta-seig", "test", evt.Content.AsMessage().Body)
					if err != nil {
						panic(err)
					}
					botClient.Client.SendText(evt.RoomID, fulfillmentText)
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

func DetectIntentText(projectID, sessionID, text string) (string, error) {
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
