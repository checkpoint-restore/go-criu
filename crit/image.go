package crit

import (
	"encoding/json"
	"fmt"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/fdinfo"
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
	Extra    string
	Humanize bool `json:"-"`
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

	if c.Humanize {
		if fe, ok := c.Message.(*fdinfo.FileEntry); ok {
			data, err := marshalFileEntryHuman(fe)
			if err != nil {
				return nil, err
			}
			if c.Extra != "" {
				return appendExtraJSON(data, c.Extra)
			}
			return data, nil
		}
	}

	data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(c.Message)
	if err != nil {
		return nil, err
	}
	// Append extra
	if c.Extra != "" {
		return appendExtraJSON(data, c.Extra)
	}
	return data, nil
}

func appendExtraJSON(data []byte, extra string) ([]byte, error) {
	extraJSON, err := json.Marshal(extra)
	if err != nil {
		return nil, err
	}
	data[len(data)-1] = byte(',')
	data = append(data, []byte(`"extra":`)...)
	data = append(data, extraJSON...)
	data = append(data, '}')
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
func splitJSONData(data []byte) ([]byte, string, error) {
	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, "", err
	}
	extraJSON, ok := fields["extra"]
	if !ok {
		return data, "", nil
	}
	var extra string
	if err := json.Unmarshal(extraJSON, &extra); err != nil {
		return nil, "", fmt.Errorf("invalid extra payload: %w", err)
	}
	delete(fields, "extra")
	payload, err := json.Marshal(fields)
	if err != nil {
		return nil, "", err
	}
	return payload, extra, nil
}

// unmarshalDefault is used for all JSON data
// that is in the standard protobuf format
func unmarshalDefault(imgData *jsonImage, img *CriuImage) error {
	for _, data := range imgData.JSONEntries {
		// Create proto struct to hold payload
		payload := proto.Clone(img.EntryType)
		jsonPayload, extraPayload, err := splitJSONData(data)
		if err != nil {
			return err
		}
		if img.Magic == "FILES" {
			var err error
			jsonPayload, err = normalizeFileEntryJSON(jsonPayload)
			if err != nil {
				return err
			}
		}
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
	if len(imgData.JSONEntries) == 0 {
		return fmt.Errorf("ghost image has no primary entry")
	}
	// Process primary entry
	entry := &CriuEntry{Message: &ghost_file.GhostFileEntry{}}
	jsonPayload, extraPayload, err := splitJSONData(imgData.JSONEntries[0])
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(jsonPayload, entry.Message); err != nil {
		return err
	}
	entry.Extra = extraPayload
	img.Entries = append(img.Entries, entry)
	// If there is only one JSON entry,
	// then no ghost chunks are present
	if len(imgData.JSONEntries) == 1 {
		return nil
	}

	// Process chunks
	for _, data := range imgData.JSONEntries[1:] {
		entry := &CriuEntry{Message: &ghost_file.GhostChunkEntry{}}
		jsonPayload, extraPayload, err = splitJSONData(data)
		if err != nil {
			return err
		}
		if err := protojson.Unmarshal(jsonPayload, entry.Message); err != nil {
			return err
		}
		entry.Extra = extraPayload
		img.Entries = append(img.Entries, entry)
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

// Helper to enable humanized output for image entries
func applyHumanize(img *CriuImage, humanize bool) {
	if img == nil || !humanize {
		return
	}
	for _, entry := range img.Entries {
		entry.Humanize = true
	}
}
