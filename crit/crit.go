package crit

import (
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
	Decode() (*CriuImage, error)
	Encode() (*CriuImage, error)
	Info() (*CriuImage, error)
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
		return nil, errors.New(fmt.Sprint("Error opening image file: ", err))
	}
	defer imgFile.Close()
	// Convert binary image to Go struct
	return decodeImg(imgFile, c.noPayload)
}

func (c *crit) Info() (*CriuImage, error) {
	imgFile, err := os.Open(c.inputFilePath)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error opening image file: ", err))
	}
	defer imgFile.Close()
	// Convert binary image to Go struct
	return countImg(imgFile)
}

func (c *crit) Encode() (*CriuImage, error) { return nil, nil }
func (c *crit) X() error                    { return nil }
