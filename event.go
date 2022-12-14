package rule

type EventType int32

const (
	EventTypeUpdate EventType = 0
	EventTypeDelete EventType = 1
)
