package main

import (
	"controller"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"model"
	"net/http"
	"os"
	"path"
)

type Parser interface {
	ParseJSON([]byte) error
}

func load(configFile string, p Parser) {
	var err error
	var input = io.ReadCloser(os.Stdin)
	if input, err = os.Open(configFile); err != nil {
		panic(err)
	}

	// Read the config file
	jsonBytes, err := ioutil.ReadAll(input)
	input.Close()
	if err != nil {
		panic(err)
	}

	// Parse the config
	if err := p.ParseJSON(jsonBytes); err != nil {
		panic(err)
	}
}

func main() {
	load("config.json", config)
	model.Configure(config.Database)

	controller.Configure(path.Join(path.Dir(os.Args[0]), config.Templates))

	err := http.ListenAndServeTLS(config.Host, config.Certificate.CertFile, config.Certificate.KeyFile, nil)
	if err != nil {
		panic(err)
	}
}

// *****************************************************************************
// Application Settings
// *****************************************************************************

// config the settings variable
var config = &Configuration{}

// configuration contains the application settings
type Configuration struct {
	Database    model.DatabaseInfo `json:"Database"`
	Certificate CertificateInfo    `json:"Certificate"`
	Templates   string             `json:"Templates"`
	Host        string             `json:"Host"`
}

type CertificateInfo struct {
	CertFile string
	KeyFile  string
}

// ParseJSON unmarshals bytes to structs
func (c *Configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}
