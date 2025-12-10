package crit

import (
	"encoding/json"
	"fmt"
	"strings"

	ghost_file "github.com/checkpoint-restore/go-criu/v8/crit/images/ghost-file"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pagemap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// CriuImage represents a CRIU binary image file
type CriuImage struct {
	Magic     string        `json:"magic"`
	Entries   []*CriuEntry  `json:"entries"`
	EntryType proto.Message `json:"-"`
}

// CriuEntry represents a single entry in an image
type CriuEntry struct {
	proto.Message
	Extra string
}

// MarshalJSON is the marshaler for CriuEntry.
// This is required as protojson.Marshal is
// used for the proto.Message, and any extra
// data is manually appended to the entry
func (c *CriuEntry) MarshalJSON() ([]byte, error) {
	// Special handling for "count"
	if c.Message == nil {
		return []byte(fmt.Sprint(`{"count":"`, c.Extra, `"}`)), nil
	}

	data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(c.Message)
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

// jsonImage is a temporary struct to store all
// entries as raw JSON, and unmarshal them into
// proper proto structs depending on the magic
type jsonImage struct {
	Magic       string            `json:"magic"`
	JSONEntries []json.RawMessage `json:"entries"`
}

// UnmarshalJSON is the unmarshaler for CriuImage.
// This is required as the object must be checked
// for any extra data, which must be removed from
// the JSON byte stream before unmarshaling the
// remaining bytes into a proto.Message object
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
func splitJSONData(data []byte) ([]byte, string) {
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

// unmarshalDefault is used for all JSON data
// that is in the standard protobuf format
func unmarshalDefault(imgData *jsonImage, img *CriuImage) error {
	for _, data := range imgData.JSONEntries {
		// Create proto struct to hold payload
		payload := proto.Clone(img.EntryType)
		jsonPayload, extraPayload := splitJSONData(data)
		// Handle proto data
		if err := protojson.Unmarshal(jsonPayload, payload); err != nil {
			return err
		}
		img.Entries = append(img.Entries, &CriuEntry{
			Message: payload,
			Extra:   extraPayload,
		})
	}

	return nil
}

// Special handler for ghost image
func unmarshalGhostFile(imgData *jsonImage, img *CriuImage) error {
	// Process primary entry
	entry := CriuEntry{Message: &ghost_file.GhostFileEntry{}}
	jsonPayload, extraPayload := splitJSONData(imgData.JSONEntries[0])
	if err := protojson.Unmarshal(jsonPayload, entry.Message); err != nil {
		return err
	}
	entry.Extra = extraPayload
	img.Entries = append(img.Entries, &entry)
	// If there is only one JSON entry,
	// then no ghost chunks are present
	if len(imgData.JSONEntries) == 1 {
		return nil
	}

	// Process chunks
	for _, data := range imgData.JSONEntries[1:] {
		entry = CriuEntry{Message: &ghost_file.GhostChunkEntry{}}
		jsonPayload, extraPayload = splitJSONData(data)
		if err := protojson.Unmarshal(jsonPayload, entry.Message); err != nil {
			return err
		}
		entry.Extra = extraPayload
		img.Entries = append(img.Entries, &entry)
	}

	return nil
}

// Special handler for pagemap image
func unmarshalPagemap(imgData *jsonImage, img *CriuImage) error {
	// First entry is pagemap head
	var payload proto.Message = &pagemap.PagemapHead{}
	for _, data := range imgData.JSONEntries {
		entry := CriuEntry{Message: payload}
		if err := protojson.Unmarshal(data, entry.Message); err != nil {
			return err
		}
		img.Entries = append(img.Entries, &entry)
		// Create struct for next entry
		payload = &pagemap.PagemapEntry{}
	}

	return nil
}
