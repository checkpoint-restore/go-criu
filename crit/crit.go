package crit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type crit struct {
	inputFilePath  string
	outputFilePath string
	inputDirPath   string
	exploreType    string
	pretty         bool
	noPayload      bool
}

type CritSvc interface {
	Show() error
	Encode() error
	Decode() error
	X() error
	Info() error
}

type criuImage struct {
	Magic   string            `json:"magic"`
	Entries []json.RawMessage `json:"entries"`
}

func New(
	inputFilePath, outputFilePath,
	inputDirPath, exploreType string,
	pretty, noPayload bool,
) CritSvc {
	return &crit{
		inputFilePath:  inputFilePath,
		outputFilePath: outputFilePath,
		inputDirPath:   inputDirPath,
		exploreType:    exploreType,
		pretty:         pretty,
		noPayload:      noPayload,
	}
}

func (c *crit) Decode() error {
	imgFile, err := os.Open(c.inputFilePath)
	if err != nil {
		return errors.New(fmt.Sprint("Error opening image file: ", err))
	}
	defer imgFile.Close()

	// Convert binary image to Go struct
	img, err := loadImg(imgFile, c.noPayload)
	if err != nil {
		return errors.New(fmt.Sprint("Error processing binary file: ", err))
	}

	var jsonData []byte
	if c.pretty {
		jsonData, err = json.MarshalIndent(img, "", "    ")
	} else {
		jsonData, err = json.Marshal(img)
	}
	if err != nil {
		return errors.New(fmt.Sprint("Error processing data into JSON: ", err))
	}
	// If no output file, print to stdout
	if c.outputFilePath == "" {
		fmt.Println(string(jsonData))
		return nil
	}
	// Write to output file
	jsonFile, err := os.Create(c.outputFilePath)
	if err != nil {
		return errors.New(fmt.Sprint("Error opening destination file: ", err))
	}
	defer jsonFile.Close()
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return errors.New(fmt.Sprint("Error writing JSON data: ", err))
	}
	return nil
}

func (c *crit) Show() error   { return nil }
func (c *crit) Encode() error { return nil }
func (c *crit) X() error      { return nil }
func (c *crit) Info() error   { return nil }
