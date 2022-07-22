package crit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// CritSvc is the interface that wraps all CRIT operations.
// To create a CRIT service instance, use New().
type CritSvc interface {
	// Read binary image file into Go struct (decode.go)
	Decode() (*CriuImage, error)
	// Read only counts of image file entries into Go struct
	Info() (*CriuImage, error)
	// Read JSON into Go struct
	Parse() (*CriuImage, error)
	// Write JSON to binary image file (encode.go)
	Encode(*CriuImage) error
	// Explore process information (explore.go)
	ExplorePs() (*PsTree, error)
	ExploreFds() ([]*Fd, error)
	ExploreMems() ([]*MemMap, error)
	ExploreRss() ([]*RssMap, error)
}

// crit implements the CritSvc interface. It contains:
// * Path of the input file
// * Path of the output file
// * Path of the input directory (for `crit explore`)
// * Boolean to provide indented and multi-line JSON output
// * Boolean to skip payload data
// * Boolean to indicate CLI usage
type crit struct {
	inputFilePath  string
	outputFilePath string
	// Directory path is required only for exploring
	inputDirPath string
	pretty       bool
	noPayload    bool
	cli          bool
}

// New creates a CRIT service to use in a Go program
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

// Decode loads a binary image file into a CriuImage object
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

// Info loads a binary image file into a CriuImage object
// with a single entry - the number of entries in the file.
// No payload data is present in the returned object.
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

// Parse is the JSON equivalent of Decode.
// It loads a JSON file into a CriuImage object.
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

// Encode dumps a CriuImage object into a binary image file
func (c *crit) Encode(img *CriuImage) error {
	// If no output path is provided in the CLI, print to stdout
	if c.outputFilePath == "" {
		if c.cli {
			return encodeImg(img, os.Stdout)
		}
	}
	imgFile, err := os.Create(c.outputFilePath)
	if err != nil {
		return errors.New(fmt.Sprint("Error opening destination file: ", err))
	}
	defer imgFile.Close()
	// Convert JSON to Go struct
	return encodeImg(img, imgFile)
}
