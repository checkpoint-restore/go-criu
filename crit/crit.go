package crit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type crit struct {
	inputFilePath  string
	outputFilePath string
	// Directory path is required only for `crit explore`
	inputDirPath string
	pretty       bool
	noPayload    bool
	cli          bool
}

type CritSvc interface {
	// Read binary file into Go struct (`decode.go`)
	Decode() (*CriuImage, error)
	// Read only counts of binary file entries into Go struct
	Info() (*CriuImage, error)
	// Read JSON into Go struct
	Parse() (*CriuImage, error)
	// Write JSON to binary file (`encode.go`)
	Encode(*CriuImage) error
	// Explore process information (`explore.go`)
	ExplorePs() (*PsTree, error)
	ExploreFds() ([]*Fd, error)
	ExploreMems() ([]*MemMap, error)
	ExploreRss() ([]*RssMap, error)
}

func New(
	inputFilePath, outputFilePath,
	inputDirPath string,
	pretty, noPayload bool,
) CritSvc {
	return &crit{
		inputFilePath:  inputFilePath,
		outputFilePath: outputFilePath,
		inputDirPath:   inputDirPath,
		pretty:         pretty,
		noPayload:      noPayload,
		cli:            false,
	}
}

// NewCli creates a CRIT service to use in a CLI app.
// All functions called by this service will wait for
// input from stdin if an input path is not provided.
func NewCli(
	inputFilePath, outputFilePath,
	inputDirPath string,
	pretty, noPayload bool,
) CritSvc {
	return &crit{
		inputFilePath:  inputFilePath,
		outputFilePath: outputFilePath,
		inputDirPath:   inputDirPath,
		pretty:         pretty,
		noPayload:      noPayload,
		cli:            true,
	}
}

func (c *crit) Decode() (*CriuImage, error) {
	// If no input path is provided in the CLI, read
	// from stdin (pipe, redirection, or keyboard)
	if c.inputFilePath == "" {
		if c.cli {
			return decodeImg(os.Stdin, c.noPayload)
		}
	}

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
	// If no input path is provided in the CLI, read
	// from stdin (pipe, redirection, or keyboard)
	if c.inputFilePath == "" {
		if c.cli {
			return countImg(os.Stdin)
		}
	}

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
	var (
		jsonData []byte
		err      error
	)

	// If no input path is provided in the CLI, read
	// from stdin (pipe, redirection, or keyboard)
	if c.inputFilePath == "" {
		if c.cli {
			jsonData, err = io.ReadAll(os.Stdin)
		}
	} else {
		jsonData, err = os.ReadFile(c.inputFilePath)
	}

	if err != nil {
		return nil, errors.New(fmt.Sprint("Error reading JSON: ", err))
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
