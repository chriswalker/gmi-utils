/*
Package gemini implements a more-than-basic Gemini client for connecting to Gemini
servers.

See:
    https://gemini.circumlunar.space/docs/specification.gmi
or
    gemini://gemini.circumlunar.space/docs/specification.gmi

for the full specification.
*/

package gemini

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"time"
)

const (
	// Maximum allowed length of a Gemini URL
	URLMaxLen = 1024
	// Default port for Gemini servers
	defaultPort = "1965"
	// URL scheme for Gemini
	Scheme = "gemini"
	// Default MIME type for gemtext
	MIMEType = "text/gemini"
)

// Status codes for advanced clients implementing the full protocol
const (
	StatusInput               = 10
	StatusSensitiveInput      = 11
	StatusSuccess             = 20
	StatusRedirectTemporary   = 30
	StatusRedirectPermanent   = 31
	StatusTemporaryFailure    = 40
	StatusServerUnavailable   = 41
	StatusCGIError            = 42
	StatusProxyError          = 43
	StatusSlowDown            = 44
	StatusPermanentFailure    = 50
	StatusNotFound            = 51
	StatusGone                = 52
	StatusProxyRequestRefused = 53
	StatusBadRequest          = 59
	StatusClientCertRequired  = 60
	StatusCertNotAuthorised   = 61
	StatusCertNotValid        = 62
)

var statusText = map[int]string{
	StatusInput:               "Input",
	StatusSensitiveInput:      "Sensitive Input",
	StatusSuccess:             "Success",
	StatusRedirectTemporary:   "Temporary Redirect",
	StatusRedirectPermanent:   "Permanent Redirect",
	StatusTemporaryFailure:    "Temporary Failure",
	StatusServerUnavailable:   "Server Unavailable",
	StatusCGIError:            "CGI Error",
	StatusProxyError:          "Proxy Error",
	StatusSlowDown:            "Slow Down",
	StatusPermanentFailure:    "Permanent Failure",
	StatusNotFound:            "Not Found",
	StatusGone:                "Gone",
	StatusProxyRequestRefused: "Proxy Request Refused",
	StatusBadRequest:          "Bad Request",
	StatusClientCertRequired:  "Client Certificate Required",
	StatusCertNotAuthorised:   "Certificate Not Authorised",
	StatusCertNotValid:        "Certificate Not Valid",
}

// StatusText returns the textual representation of the supplied
// Status Code.
func StatusText(code int) string {
	return statusText[code]
}

// Status returns a formatted string for the supplied Status Code,
// comprised of the Status Code and the textual representation of
// the code - e.g. "20 (Success)", "60 (Client Certificate Required)".
func Status(code int) string {
	return fmt.Sprintf("%d (%s)", code, statusText[code])
}

// Response encapsulates a Gemini response.
type Response struct {
	// ResponseDuration records the time it took
	// to receive a response
	ResponseDuration time.Duration

	// URL is the URL used to obtain this response.
	URL url.URL

	// StatusCode is the response status code.
	StatusCode int

	// Meta holds any response header meta values; these
	// vary depending on status code.
	Meta string

	// ContentLength records the length of the received
	// content.
	ContentLength int

	// Body is an array of bytes holding the received
	// page content.
	Body []byte
}

// DefaultClient is a barebones client, used for basic Gemini calls. It
// currently defaults to skipping cert verification. Applications requiring
// this should create their own client with appropriate TLS configuration.
var DefaultClient = NewClient(
	Timeout(9*time.Second),
	Config(&tls.Config{InsecureSkipVerify: true}),
)

// Get retrieves the given Gemini resource. It is a convenient function
// allowing callers to get a Gemini page without having to create and
// manage clients, It uses the DefaultClient under the hood, and so
// no TLS cert checking is performed.
func Get(url string) (*Response, error) {
	return DefaultClient.Get(url)
}
