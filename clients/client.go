package clients

import (
	"github.com/locmai/yuta-messaging/config"
	"maunium.net/go/mautrix"
)

func NewMatrixClient(c config.ClientConfig) (MatrixClient, error) {
	client, err := mautrix.NewClient(c.HomeserverURL, "", "")
	if err != nil {
		panic(err)
	}

	authType := mautrix.AuthTypePassword
	if c.AccessToken != "" {
		authType = mautrix.AuthTypeToken
		_, err := client.Login(&mautrix.ReqLogin{
			Type:             authType,
			Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: c.Username},
			Token:            c.AccessToken,
			StoreCredentials: true,
		})
		if err != nil {
			panic(err)
		}
		return MatrixClient{Client: client}, nil
	}

	_, err = client.Login(&mautrix.ReqLogin{
		Type:             authType,
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: c.Username},
		Password:         c.Password,
		StoreCredentials: true,
	})
	if err != nil {
		panic(err)
	}
	return MatrixClient{Client: client}, nil
}
