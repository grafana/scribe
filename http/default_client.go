package swhttp

import (
	"net/http"
	"time"
)

var DefaultClient = http.Client{
	Timeout: time.Minute * 5,
}
