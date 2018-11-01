package websupport

import (
	"fmt"
	"log"

	"github.com/radoslavoleksak/go-websupport/websupport"
)

type Config struct {
	Username     string
	Password     string
}

// Client represents the Websupport provider client.
// This is a convenient container for the configuration and the underlying API client.
type Client struct {
	client *websupport.Client
	loggedUser *websupport.User
	config *Config
}

// Client() returns a new client for accessing websupport.
func (c *Config) Client() (*Client, error) {
	client, err := websupport.NewClient(c.Username, c.Password, nil)

	if err != nil {
		return nil, fmt.Errorf("Error setting up Websupport client: %s", err)
	}

	listAllUsersResp, err := client.Users.ListAllUsers()
	if err != nil {
		return nil, fmt.Errorf("Error in Websupport client while listing all users: %s", err)
	}

	var loggedUser *websupport.User // TODO ma to byt ako pointer? stale som si neisty
	for _, user := range listAllUsersResp.Items {
		if user.Login == c.Username {
			loggedUser = &user // TODO dobre je ten zavinac? :P
		}
	}

	if loggedUser == nil {
		return nil, fmt.Errorf("Error in Websupport client. Cannot find user with Login attribute by Username: %s", c.Username)
	}

	provider := &Client{
		client: client,
		loggedUser: loggedUser,
		config: c,
	}

	log.Printf("[INFO] Websupport Client configured for username (userId: %d): %s", c.Username, loggedUser.Id)

	return provider, nil
}