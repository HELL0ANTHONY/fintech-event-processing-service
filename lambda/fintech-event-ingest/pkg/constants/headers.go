package constants

// Headers returns a map of standard HTTP headers for responses.
func Headers() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Access-Control-Allow-Origin, Access-Control-Allow-Methods, Content-Type",
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Methods": "POST, OPTIONS",
	}
}
