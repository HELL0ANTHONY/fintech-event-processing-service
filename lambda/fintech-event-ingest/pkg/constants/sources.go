package constants

// Source represents the origin of the event ingestion request.
type Source string

// Represents the different valid sources for event ingestion.
const (
	SourceAPI         Source = "api"
	SourceBatch       Source = "batch"
	SourceEventBridge        = "eventbridge"
	SourceReplay             = "replay"
	SourceTest               = "test"
)

// IsValidSource checks if the provided source is a valid Source.
func IsValidSource(source Source) bool {
	switch source {
	case SourceAPI, SourceBatch, SourceEventBridge, SourceReplay, SourceTest:
		return true
	default:
		return false
	}
}
