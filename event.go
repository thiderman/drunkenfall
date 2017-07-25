package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// An Event represents something altering the state of a tournament
//
// Kind should be a snake_case string that categorizes the event.
// Items should be a list where the odd elements are string keys and the even
// elements are values.
// Message should be a string containing "{key}" formatting strings for
// interpolation of the Items.
type Event struct {
	Date    time.Time              `json:"time"`
	Kind    string                 `json:"kind"`
	Items   map[string]interface{} `json:"items"`
	Message string                 `json:"message"`
}

// NewEvent creates a new event and maps the argument into the items
func NewEvent(kind, message string, items ...interface{}) (*Event, error) {
	if len(items)%2 != 0 {
		return nil, fmt.Errorf("Creating event with uneven items: %s", items)
	}

	m := make(map[string]interface{})
	for x := 0; x < len(items); x += 2 {
		key := items[x].(string)
		val := items[x+1]
		m[key] = val
	}

	if _, ok := m["person"]; !ok {
		return nil, fmt.Errorf("%s does not contain a person", kind)
	}

	e := Event{
		Date:    time.Now(),
		Kind:    kind,
		Message: message,
		Items:   m,
	}

	return &e, nil
}

// JSON formats the event as a JSON blob
func (e *Event) JSON() (out []byte, err error) {
	out, err = json.Marshal(e)
	return
}

func (e *Event) String() string {
	// TODO(thiderman): It would be neat to have something that just properly
	// formats the message, but for now just do something stupid and simple
	return fmt.Sprintf("<%s: %v>", e.Message, e.Items)
}
