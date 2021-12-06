package internal

type Action struct {
	Request    string
	Direction  string
	ApiVersion int16
	Message    interface{}
}
