package utility

import (
	"net/http"
	"log"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

const (
	MULTI_STORE_EXTENSION = "MultiStoreSupport"
	IN_STORE_PROCESSING_EXTENSION = "InStoreProcessing"
)
	
type ExtensionResponse struct {
	Code int
	Status string
	Data map[string]Extension
}

type Extension struct {
	Id int
	Name string
	Status string
}

type Config map[string]string

func Configure(config Config) {
	Urls = config
}

var Urls Config

func IsExtensionEnabled(organizationId int, extensionSlug string) bool {
	var url = Urls["account-service"]
	url = url + fmt.Sprintf("/extension?organizationId=%d&slug=%s",organizationId, extensionSlug)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		return false
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var response ExtensionResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("in response ExtensionResponse", response.Data["extension"].Status)
	if response.Data["extension"].Status == "ENABLED" {
		return true
	}
	return false
}