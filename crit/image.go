package crit

import (
	"fmt"
	"strings"

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

/* Custom marshaler for the entry, as we need to use
protojson.Marshal for the proto.Message, and manually
append the extra data (if present) */
func (c *CriuEntry) MarshalJSON() ([]byte, error) {
	// Special handling for "count"
	if c.Message == nil {
		return []byte(fmt.Sprint(`{"count":"`, c.Extra, `"}`)), nil
	}

	j, err := protojson.Marshal(c.Message)
	if err != nil {
		return nil, err
	}
	// Append extra
	if c.Extra != "" {
		extraString := fmt.Sprint(`"extra":"`, c.Extra, `"}`)
		j[len(j)-1] = byte(',')
		j = append(j, []byte(extraString)...)
	}
	return j, nil
}

/* Custom unmarshaler to check if extra data is present,
remove it from the JSON and unmarshal the remaining with
protojson.Unmarshal into proto.Message */
func (c *CriuEntry) UnmarshalJSON(j []byte) error {
	jString := string(j)
	jItems := strings.Split(jString, ",")
	// Check if extra data exists
	last := strings.Split(jItems[len(jItems)-1], ":")
	if last[0] == `"extra"` {
		extra := last[1]
		c.Extra = extra[1 : len(extra)-2]
		jString = strings.Join(jItems[:len(jItems)-1], ",") + "}"
	}
	err := protojson.Unmarshal([]byte(jString), c.Message)
	return err
}
