package crit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
)

// Critter is the interface that wraps all CRIT operations.
// To create a CRIT service instance, use New().
type Critter interface {
	// Read binary image file into Go struct (decode.go)
	Decode(proto.Message) (*CriuImage, error)
	// Read only counts of image file entries into Go struct
	Info() (*CriuImage, error)
	// Read JSON into Go struct
	Parse(proto.Message) (*CriuImage, error)
	// Write JSON to binary image file (encode.go)
	Encode(*CriuImage) error
	// Explore process information (explore.go)
	ExplorePs() (*PsTree, error)
	ExploreFds() ([]*Fd, error)
	ExploreMems() ([]*MemMap, error)
	ExploreRss() ([]*RssMap, error)
	ExploreSk() ([]*Sk, error)
}

// crit implements the Critter interface. It contains:
// * Path of the input file
// * Path of the output file
// * Path of the input directory (for `crit explore`)
// * Boolean to format and indent JSON output
// * Boolean to skip payload data
type crit struct {
	inputFile  *os.File
	outputFile *os.File
	// Directory path is required only for exploring
	inputDirPath string
	pretty       bool
	noPayload    bool
}

// New creates an instance of the CRIT service
func New(
	inputFilePath, outputFilePath *os.File,
	inputDirPath string,
	pretty, noPayload bool,
) Critter {
	return &crit{
		inputFile:    inputFilePath,
		outputFile:   outputFilePath,
		inputDirPath: inputDirPath,
		pretty:       pretty,
		noPayload:    noPayload,
	}
}

// Decode loads a binary image file into a CriuImage object
func (c *crit) Decode(entryType proto.Message) (*CriuImage, error) {
	// Convert binary image to Go struct
	return decodeImg(c.inputFile, entryType, c.noPayload)
}

// Info loads a binary image file into a CriuImage object
// with a single entry - the number of entries in the file.
// No payload data is present in the returned object.
func (c *crit) Info() (*CriuImage, error) {
	// Convert binary image to Go struct
	return countImg(c.inputFile)
}

// Parse is the JSON equivalent of Decode.
// It loads a JSON file into a CriuImage object.
func (c *crit) Parse(entryType proto.Message) (*CriuImage, error) {
	jsonData, err := io.ReadAll(c.inputFile)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON: %w", err)
	}

	img := CriuImage{EntryType: entryType}
	if err = json.Unmarshal(jsonData, &img); err != nil {
		return nil, fmt.Errorf("error processing JSON: %w", err)
	}

	return &img, nil
}

// Encode dumps a CriuImage object into a binary image file
func (c *crit) Encode(img *CriuImage) error {
	// Convert JSON to Go struct
	return encodeImg(img, c.outputFile)
}
