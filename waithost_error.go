package waithost

// Return codes
type ErrorCode uint8

//
type WaitHostError struct {
	error ErrorCode
}

const (
	TIMEOUT ErrorCode = iota
	BAD_URL
	BAD_SCHEME
	BAD_HOST
	BAD_PORT
)

func (e *WaitHostError) Error() string {
	switch e.error {
	case TIMEOUT:
		return "Timeout"
	case BAD_URL:
		return "Bad url"
	case BAD_SCHEME:
		return "Bad scheme"
	case BAD_HOST:
		return "Bad host"
	case BAD_PORT:
		return "Bad port"
	default:
		return "Bad url"
	}
}

func (e *WaitHostError) Code() ErrorCode {
	return e.error
}
