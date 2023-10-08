package annotations

import "time"

type Annotation struct {
	Text string    `json:"text"`
	Time time.Time `json:"time"`
	Tags []string  `json:"tags"`
}
