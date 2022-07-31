package crit

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/checkpoint-restore/go-criu/v5/crit/images"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type CriuImage struct {
	Magic   string       `json:"magic"`
	Entries []*CriuEntry `json:"entries"`
}

type CriuEntry struct {
	proto.Message
	Extra string
}

// Custom marshaler for the entry, as we must use
// protojson.Marshal for the proto.Message, and
// manually append any extra data to the entry
func (c *CriuEntry) MarshalJSON() ([]byte, error) {
	// Special handling for "count"
	if c.Message == nil {
		return []byte(fmt.Sprint(`{"count":"`, c.Extra, `"}`)), nil
	}

	data, err := protojson.Marshal(c.Message)
	if err != nil {
		return nil, err
	}
	// Append extra
	if c.Extra != "" {
		extraString := fmt.Sprint(`"extra":"`, c.Extra, `"}`)
		data[len(data)-1] = byte(',')
		data = append(data, []byte(extraString)...)
	}
	return data, nil
}

// Use a temporary struct to store all entries
// as raw JSON, and unmarshal them into proper
// proto structs depending on the magic
type jsonImage struct {
	Magic       string            `json:"magic"`
	JsonEntries []json.RawMessage `json:"entries"`
}

// Custom unmarshaler for the image, as we must check if any
// extra data is present, remove it from the JSON byte stream
// and unmarshal the remaining with into a proto struct
func (img *CriuImage) UnmarshalJSON(data []byte) error {
	imgData := jsonImage{}
	var err error

	if err = json.Unmarshal(data, &imgData); err != nil {
		return err
	}
	img.Magic = imgData.Magic

	switch img.Magic {
	case "GHOST_FILE":
		err = unmarshalGhostFile(&imgData, img)
	case "PAGEMAP":
		err = unmarshalPagemap(&imgData, img)
	default:
		err = unmarshalDefault(&imgData, img)
	}

	return err
}

// Helper to separate proto data and extra data
func splitJsonData(data []byte) ([]byte, string) {
	extraPayload := ""
	dataString := string(data)
	dataItems := strings.Split(dataString, ",")
	// Handle extra data, if present
	last := strings.Split(dataItems[len(dataItems)-1], ":")
	if last[0] == `"extra"` {
		extra := last[1]
		extraPayload = extra[1 : len(extra)-2]
		dataString = strings.Join(dataItems[:len(dataItems)-1], ",") + "}"
	}
	return []byte(dataString), extraPayload
}

func unmarshalDefault(imgData *jsonImage, img *CriuImage) error {
	for _, data := range imgData.JsonEntries {
		// Create proto struct to hold payload
		payload, err := images.ProtoHandler(img.Magic)
		if err != nil {
			return err
		}
		jsonPayload, extraPayload := splitJsonData(data)
		// Handle proto data
		if err = protojson.Unmarshal(jsonPayload, payload); err != nil {
			return err
		}
		img.Entries = append(img.Entries, &CriuEntry{
			Message: payload,
			Extra:   extraPayload,
		})
	}

	return nil
}

func unmarshalGhostFile(imgData *jsonImage, img *CriuImage) error {
	// Process primary entry
	entry := CriuEntry{Message: &images.GhostFileEntry{}}
	jsonPayload, extraPayload := splitJsonData(imgData.JsonEntries[0])
	if err := protojson.Unmarshal(jsonPayload, entry.Message); err != nil {
		return err
	}
	entry.Extra = extraPayload
	img.Entries = append(img.Entries, &entry)
	// If there is only one JSON entry,
	// then no ghost chunks are present
	if len(imgData.JsonEntries) == 1 {
		return nil
	}

	// Process chunks
	for _, data := range imgData.JsonEntries[1:] {
		entry = CriuEntry{Message: &images.GhostChunkEntry{}}
		jsonPayload, extraPayload = splitJsonData(data)
		if err := protojson.Unmarshal(jsonPayload, entry.Message); err != nil {
			return err
		}
		entry.Extra = extraPayload
		img.Entries = append(img.Entries, &entry)
	}

	return nil
}

func unmarshalPagemap(imgData *jsonImage, img *CriuImage) error {
	// First entry is pagemap head
	var payload proto.Message = &images.PagemapHead{}
	for _, data := range imgData.JsonEntries {
		entry := CriuEntry{Message: payload}
		if err := protojson.Unmarshal(data, entry.Message); err != nil {
			return err
		}
		img.Entries = append(img.Entries, &entry)
		// Create struct for next entry
		payload = &images.PagemapEntry{}
	}

	return nil
}
