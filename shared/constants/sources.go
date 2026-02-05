package constants

// Source represents the origin of the event ingestion request.
type Source string

const (
	SourceAPI         Source = "api"
	SourceBatch       Source = "batch"
	SourceEventBridge Source = "eventbridge"
	SourceReplay      Source = "replay"
	SourceTest        Source = "test"
)

// IsValidSource checks if the provided source is valid.
func IsValidSource(s Source) bool {
	switch s {
	case SourceAPI, SourceBatch, SourceEventBridge, SourceReplay, SourceTest:
		return true
	default:
		return false
	}
}
