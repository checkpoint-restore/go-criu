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
	// Read binary file into Go struct
	Decode() (*CriuImage, error)
	// Read only counts of binary file entries into Go struct
	Info() (*CriuImage, error)
	// Read JSON into Go struct
	Parse() (*CriuImage, error)
	// Write JSON to binary file
	Encode(*CriuImage) error
	// Explore process information
	X() error
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

func (c *crit) Decode() (*CriuImage, error) {
	imgFile, err := os.Open(c.inputFilePath)
	if err != nil {
		return nil,
			errors.New(fmt.Sprint("Error opening image file: ", err))
	}
	defer imgFile.Close()
	// Convert binary image to Go struct
	return decodeImg(imgFile, c.noPayload)
}

func (c *crit) Info() (*CriuImage, error) {
	imgFile, err := os.Open(c.inputFilePath)
	if err != nil {
		return nil,
			errors.New(fmt.Sprint("Error opening image file: ", err))
	}
	defer imgFile.Close()
	// Convert binary image to Go struct
	return countImg(imgFile)
}

func (c *crit) Parse() (*CriuImage, error) {
	jsonData, err := os.ReadFile(c.inputFilePath)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error opening JSON file: ", err))
	}

	img := CriuImage{}
	if err = json.Unmarshal(jsonData, &img); err != nil {
		return nil, errors.New(fmt.Sprint("Error processing JSON: ", err))
	}

	return &img, nil
}

func (c *crit) Encode(img *CriuImage) error {
	imgFile, err := os.Create(c.outputFilePath)
	if err != nil {
		return errors.New(fmt.Sprint("Error opening destination file: ", err))
	}
	defer imgFile.Close()
	// Convert JSON to Go struct
	return encodeImg(img, imgFile)
}

func (c *crit) X() error { return nil }
