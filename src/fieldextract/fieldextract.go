package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/karamani/iostreams"
)

var (
	debugMode     bool
	fieldsArg     string
	formatArg     string
	separatorArg  string
	withNamesArg  bool
	skipEmptyArg  bool
	trimQuotesArg bool
)

func main() {
	app := cli.NewApp()
	app.Name = "fieldextract"
	app.Usage = "Retrieves the fields of data structures & prints them to stdout"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug mode",
			Destination: &debugMode,
		},
		cli.StringFlag{
			Name:        "fields",
			Usage:       "Fields to extract",
			Destination: &fieldsArg,
		},
		cli.StringFlag{
			Name:        "format",
			Usage:       "Format of data structure",
			Value:       "json",
			Destination: &formatArg,
		},
		cli.StringFlag{
			Name:        "separator",
			Usage:       "Output separator",
			Value:       "\t",
			Destination: &separatorArg,
		},
		cli.BoolFlag{
			Name:        "withnames",
			Usage:       "Output field names",
			Destination: &withNamesArg,
		},
		cli.BoolFlag{
			Name:        "trimquotes",
			Usage:       "trim quotes",
			Destination: &trimQuotesArg,
		},
		cli.BoolFlag{
			Name:        "skipempty",
			Usage:       "skip rows with empty fields",
			Destination: &skipEmptyArg,
		},
	}

	app.Action = func(c *cli.Context) {

		// this func's called for each row in stdinы
		process := func(row []byte) error {

			debug(string(row))

			res, err := extractFields(string(row), fieldsArg)
			if err != nil {
				log.Printf("ERROR: %s\n", err.Error())
			} else {
				fmt.Println(res)
			}

			return nil
		}

		err := iostreams.ProcessStdin(process)
		if err != nil {
			log.Panicln(err.Error())
		}
	}

	app.Run(os.Args)
}

func debug(msg string) {
	if debugMode {
		fmt.Println("DEBUG: " + msg)
	}
}

func extractFields(jsonString, fieldsString string) (string, error) {

	var objmap map[string]*json.RawMessage

	err := json.Unmarshal([]byte(jsonString), &objmap)
	if err != nil {
		return "", err
	}

	res := ""

	fields := strings.Split(fieldsString, ",")
	for i, oneField := range fields {
		fieldParts := strings.Split(oneField, ".")
		fieldValue, err := extractOneField(objmap, fieldParts)
		if err != nil {
			log.Println("ERROR: " + err.Error())
			fieldValue = ""
		}
		if skipEmptyArg && len(fieldValue) == 0 {
			return "", errors.New("Empty field " + oneField)
		}
		if i > 0 {
			res += separatorArg
		}
		if withNamesArg {
			res += oneField + ":"
		}
		if trimQuotesArg {
			fieldValue = strings.Trim(fieldValue, "\"")
		}
		res += fieldValue
	}

	return res, nil
}

func extractOneField(objmap map[string]*json.RawMessage, fieldParts []string) (string, error) {

	var innerObjMap map[string]*json.RawMessage = objmap

	lastIndex := len(fieldParts) - 1
	for i, part := range fieldParts {

		obj, ok := innerObjMap[part]
		if !ok || obj == nil {
			return "", nil
		}

		if i == lastIndex {
			return string(*obj), nil
		}

		err := json.Unmarshal([]byte(*obj), &innerObjMap)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}
