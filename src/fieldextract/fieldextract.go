package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"log"
	"os"
	"strings"
)

var (
	debugMode    bool
	fieldsArg    string
	formatArg    string
	separatorArg string
	withNamesArg bool
	skipEmptyArg bool
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
			Name:        "skipempty",
			Usage:       "skip rows with empty fields",
			Destination: &skipEmptyArg,
		},
	}
	app.Action = func(c *cli.Context) {

		reader := bufio.NewReader(os.Stdin)
		jsonString := ""
		for {
			bytes, hasMoreInLine, err := reader.ReadLine()
			if err != nil {
				if err != io.EOF {
					log.Fatalf("ERROR: %s\n", err.Error())
				}
				break
			}
			jsonString += string(bytes)
			if !hasMoreInLine {

				debug(jsonString)

				res, err := extractFields(jsonString, fieldsArg)
				if err != nil {
					log.Printf("ERROR: %s\n", err.Error())
				} else {
					fmt.Println(res)
				}

				jsonString = ""
			}
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
			if obj == nil {
				return "", nil
			}
			return string(*obj), nil
		}

		err := json.Unmarshal([]byte(*obj), &innerObjMap)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}
