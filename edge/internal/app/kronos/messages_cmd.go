package kronos

import (
	"devais.it/kronos/internal/pkg/sync/messages"
	"encoding/json"
	"github.com/alecthomas/jsonschema"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

func messageFromString(messageStr string) (interface{}, error) {
	var message interface{}

	switch strings.ToLower(messageStr) {
	case "connected":
		message = &messages.Connected{}
	case "disconnected":
		message = &messages.Disconnected{}
	case "event":
		message = &messages.Event{}
	case "versions":
		message = &messages.Versions{}
	case "command":
		message = &messages.ServerCommand{}
	case "command_response":
		message = &messages.CommandResponse{}
	case "sync":
		message = &messages.Sync{}
	default:
		return nil, eris.Errorf("Invalid message: '%s'", messageStr)
	}

	return message, nil
}

func writeMessageSchema(w io.Writer, message interface{}, ident int) error {
	schema := jsonschema.Reflect(message)
	bytes, err := schema.MarshalJSON()
	if err != nil {
		return err
	}

	// Reformat
	if ident > 0 {
		var jsonData interface{}
		err = json.Unmarshal(bytes, &jsonData)
		if err != nil {
			return err
		}
		bytes, err = json.MarshalIndent(
			&jsonData,
			"",
			strings.Repeat(" ", ident),
		)
		if err != nil {
			return err
		}
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

type printMessageSchemaCmd struct {
	Message string `kong:"arg,name=message,enum='connected,disconnected,event,versions,command,command_response,sync',help='Message type'"`
	Ident   int    `kong:"arg,name=ident,optional,default=0,help=JSON ident"`
}

func (c *printMessageSchemaCmd) Run(*Context) error {
	message, err := messageFromString(c.Message)
	if err != nil {
		return err
	}
	return writeMessageSchema(os.Stdout, message, c.Ident)
}

type saveMessageSchemaCmd struct {
	Filename string `kong:"arg,name=file,help='Schema file',type=file"`
	Message  string `kong:"arg,name=message,enum='connected,disconnected,event,versions,command,command_response,sync',help='Message type'"`
	Ident    int    `kong:"arg,name=ident,optional,default=0,help:'JSON ident'"`
}

func (c *saveMessageSchemaCmd) Run(*Context) error {
	file, err := os.Create(c.Filename)
	if err != nil {
		return eris.Wrapf(err, "failed to open '%s' file for creation", c.Filename)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Errorf("failed to close file '%s': %s", c.Filename, eris.ToString(err, false))
		}
	}()

	message, err := messageFromString(c.Message)
	if err != nil {
		return err
	}
	return writeMessageSchema(file, message, c.Ident)
}
