package httpcli

import (
	"net/http"
	"time"
)

var HttpClient = &http.Client{
	Timeout: time.Second * 10,
}
