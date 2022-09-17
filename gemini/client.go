package gemini

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// option configures an aspect of the Gemini client.
type option func(c *Client)

// Timeout sets a socket connect timeout option on the client.
func Timeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		c.dialer.NetDialer.Timeout = timeout
	}
}

// Config sets TLS Config for the client's TSL Dialer.
func Config(config *tls.Config) func(*Client) {
	return func(c *Client) {
		c.dialer.Config = config
	}
}

// NewClient creates an instance of the Gemini client, configured as per the
// option functions passed in.
func NewClient(opts ...option) *Client {
	c := &Client{
		dialer: new(tls.Dialer),
	}
	c.dialer.NetDialer = new(net.Dialer)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Client is a Gemini client complete with configured
// TLS dialer.
type Client struct {
	dialer *tls.Dialer
}

// Get attempts to get the supplied Gemini URL.
func (c *Client) Get(geminiURL string) (*Response, error) {
	// If scheme missing, default to gemini://
	if !strings.Contains(geminiURL, "://") {
		geminiURL = fmt.Sprintf("%s://%s", Scheme, geminiURL)
	}

	// Parse the supplied URL, and do some sanity checking on
	// it - reject anything non-Gemini
	url, err := url.ParseRequestURI(geminiURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing supplied URL: %w", err)
	}
	if url.Scheme != Scheme {
		return nil, fmt.Errorf("unsupported URL scheme '%s'", url.Scheme)
	}

	return c.get(*url)
}

// buildHostString takes the supplied URL and extracts the relevant
// components for constructing a TCP host to connect to.
func buildHostString(url url.URL) (string, error) {
	port := defaultPort
	if url.Port() != "" {
		port = url.Port()
	}

	return fmt.Sprintf("%s:%s", url.Hostname(), port), nil
}

// getConn connects to the given Gemini server and returns
// the resulting net.Conn.
func (c *Client) getConn(url url.URL) (net.Conn, error) {
	hostStr, err := buildHostString(url)
	if err != nil {
		return nil, err
	}

	// TODO: Check TLS certs here; keep in TOFU store a'la SSH's known_hosts

	return c.dialer.Dial("tcp", hostStr)
}

// get makes the actual request over the internal net.Conn.
func (c *Client) get(url url.URL) (*Response, error) {
	conn, err := c.getConn(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to '%s': %s", url.String(), err.Error())
	}
	defer conn.Close()

	start := time.Now()

	// Send request to server - composed of the URL plus CRLF
	_, err = conn.Write([]byte(url.String() + "\r\n"))
	if err != nil {
		return nil, fmt.Errorf("could not send request to server: %w", err)
	}

	// Process response
	reader := bufio.NewReader(conn)
	rspHeader, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("could not read response header: %w", err)
	}

	rsp := &Response{URL: url}
	rsp.StatusCode, rsp.Meta, err = parseHeader(rspHeader)
	if err != nil {
		return nil, fmt.Errorf("could not parse response header: %w", err)
	}

	switch rsp.StatusCode {
	case StatusInput,
		StatusSensitiveInput:
		return nil, fmt.Errorf("unsupported feature")
	case StatusSuccess:
		body, err := processResponse(rsp.Meta, reader)
		if err != nil {
			return nil, err
		}

		rsp.ContentLength = len(body)
		rsp.Body = body
		rsp.ResponseDuration = time.Since(start)
	case StatusRedirectTemporary,
		StatusRedirectPermanent:
		url, err := url.Parse(rsp.Meta)
		if err != nil {
			return nil, fmt.Errorf("error parsing redirect URL: %w", err)
		}

		return c.get(*url)
	case StatusBadRequest:
		return nil, fmt.Errorf("server could not process request: %s", rsp.Meta)
	case StatusTemporaryFailure,
		StatusServerUnavailable,
		StatusPermanentFailure:
		// TODO, temporary
		return nil, fmt.Errorf("error: %s", rsp.Meta)
	case StatusClientCertRequired:
		// Can test with gemini://astrobotany.mozz.us/app
		return nil, fmt.Errorf("resource requires a client certificate")
	case StatusCertNotAuthorised,
		StatusCertNotValid:
		// TODO, temporary
		return nil, fmt.Errorf("certificate problem: %s", rsp.Meta)
	}

	return rsp, nil
}

// parseHeader parses the response header, which should be composed
// of a status code and optional response metadata.
func parseHeader(header string) (int, string, error) {
	header = strings.Trim(header, "\r\n ")
	parts := strings.SplitN(header, " ", 2)

	status, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("could not extract response status code: %s", err.Error())
	}

	var meta string
	if len(parts) > 1 {
		meta = strings.TrimSuffix(parts[1], "\n")
	}

	return status, meta, nil
}

// processResponse reads in a successful response, returning its body.
func processResponse(meta string, reader io.Reader) ([]byte, error) {
	if !strings.HasPrefix(meta, "text/gemini") {
		return nil, fmt.Errorf("unsupported MIME type '%s'", meta)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	return body, nil
}
