package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	utils "filesender/utils"

	"github.com/woanware/gdriver"
	"github.com/woanware/gdriver/oauthhelper"
)

//
func InitialiseGoogleDrive() (*gdriver.FileInfo, *gdriver.GDriver) {

	// Setup OAuth
	helper := oauthhelper.Auth{
		ClientID:     "450649512071-cfe5626fbsh1ei1jm783pm9bmcd88fu6.apps.googleusercontent.com",
		ClientSecret: "0WRUgxWx41cohjTqHx33fdBY",
		Authenticate: func(url string) (string, error) {
			fmt.Printf("Open to authorize 'filesender' to access your drive\n%s\n", url)

			var code string
			fmt.Printf("Code: ")
			if _, err := fmt.Scan(&code); err != nil {
				return "", fmt.Errorf("Unable to read authorization code %v", err)
			}
			return code, nil
		},
	}

	var err error
	// Try to load a client token from file
	helper.Token, err = oauthhelper.LoadTokenFromFile("token.json")
	if err != nil {
		// if the error is NotExist error continue
		// we will create a token
		if !os.IsNotExist(err) {
			log.Panic(err)
		}
	}

	// Create a new authorized HTTP client
	client, err := helper.NewHTTPClient(context.Background())
	if err != nil {
		log.Panic(err)
	}

	// store the token for future use
	if err = oauthhelper.StoreTokenToFile("token.json", helper.Token); err != nil {
		log.Panic(err)
	}

	// create a gdriver instance
	gdrive, err := gdriver.New(client)
	if err != nil {
		log.Panic(err)
	}

	var dir *gdriver.FileInfo
	dir, err = gdrive.Stat("filesender")
	if err != nil {
		if strings.Contains(err.Error(), "not found") == false {
			log.Panic(err)
		} else {
			dir, err = gdrive.MakeDirectory("filesender")
			if err != nil {
				utils.OutputAndExit(fmt.Sprintf("Error creating google drive folder: %v", err))
			}
		}

	}

	return dir, gdrive
}
