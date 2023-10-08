package annotations

import "time"

type Annotation struct {
	Text string
	Time time.Time
	Tags []string
}
