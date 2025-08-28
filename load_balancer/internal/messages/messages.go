package messages

// error messages
const (
	ErrLoadConfig         = "failed to load config file"
	ErrLAS                = "listenAndServe failed"
	ErrShutdown           = "server Shutdown Failed"
	ErrAllAttemptsFailed  = "all attempts failed"
	ErrNoBackends         = "no backends are reachable"
	ErrAttemptFailed      = "attempt failed"
	ErrServiceUnavailable = "service unavailable"
	ErrResponse           = "failed to write a response"
	ErrInvalidBackendURL  = "invalid backend url"
	ErrReadConfig         = "unable to read config file: %v"
	ErrProxy              = "proxy error"
	ErrAddToken           = "failed to add a token"
	ErrUpdate             = "failed to update due to concurrent modification of IP %s: %v"
	ErrFind               = "failed to find due to concurrent modification of IP %s: %v"
	ErrInsert             = "failed to insert due to concurrent modification of IP %s: %v"
	ErrNoData             = "no data found for IP %s"
	ErrLimiter            = "rate limiter failed to process the request"
	ErrTooManyRequests    = "rate limit exceeded"
	ErrNoAvailableToken   = "no tokens are available"
	ErrGetKeys            = "failed to list Redis keys"
	ErrNoIPORVal          = "missing 'ip' or 'value' parameter"
	ErrBadValue           = "invalid 'value' parameter"
	ErrSetRate            = "failed to set rate"
	ErrSetMax             = "failed to set max tokens"
)

// info messages
const (
	InfoBalancerON         = "load Balancer is on"
	InfoGracefulStopStart  = "shutting down gracefully"
	InfoGracefulStopFinish = "server gracefully stopped"
	InfoForwardingURL      = "forwarding to"
	InfoForwardingActive   = "active"
	InfoSuccessfulProxy    = "successfully proxied to"
	InfoShutdownHealth     = "shutting down health checks"
	InfoUnreachable        = "server is unreachable"
	InfoReachable          = "server is reachable"
	InfoAddedToken         = "added a token"
	InfoTickersStopped     = "all tickers stopped due to server shutdown"
	InfoUserCreated        = "user created"
	InfoAccessGranted      = "access granted"
	InfoRateUPD            = "rate updated"
	InfoMaxUPD             = "max tokens updated"
)

// misc
const (
	URL    = "URL"
	Port   = "Port"
	Number = "Number"
	Code   = "Code"
	Status = "Status"
	IP     = "IP"
	Tokens = "tokens"
)
