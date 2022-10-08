package gemini

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestParseHeader(t *testing.T) {
	testCases := map[string]struct {
		header string
		status int
		meta   string
		errMsg string
	}{
		"basic": {
			header: "20 text/gemini",
			status: StatusSuccess,
			meta:   "text/gemini",
			errMsg: "",
		},
		"error": {
			header: "invalid header",
			errMsg: "could not extract response status code",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			status, meta, err := parseHeader(tc.header)
			if tc.errMsg != "" {
				if err == nil {
					t.Error("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("got error '%s', want '%s", err, tc.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %q:", err)
				return
			}
			if status != tc.status {
				t.Errorf("got status code of '%d', want '%d'", status, tc.status)
			}
			if meta != tc.meta{
				t.Errorf("got meta value of '%s', want '%s'", meta, tc.meta)
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	testCases := map[string]struct {
		meta     string
		body     string
		expected []byte
		errMsg   string
	}{
		"valid response": {
			meta:     "text/gemini",
			body:     "# Valid gemtext",
			expected: []byte("# Valid gemtext"),
			errMsg:   "",
		},
		"invalid mime type": {
			meta:     "text/invalid",
			body:     "",
			expected: nil,
			errMsg:   "unsupported MIME type 'text/invalid'",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteString(tc.body)
			b, err := processResponse(tc.meta, bufio.NewReader(&buf))
			if tc.errMsg != "" {
				if err == nil {
					t.Error("expected an error, but got nil")
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("got error '%s', want '%s", err, tc.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %q:", err)
			}
			if !bytes.Equal(b, tc.expected) {
				t.Errorf("got returned bytes of '%v', want '%v'", b, tc.expected)
			}
		})
	}
}

func TestBuildHostString(t *testing.T) {
	testCases := map[string]struct {
		url          url.URL
		expectedHost string
		expectedErr  string
	}{
		"domain only": {
			url:          url.URL{Scheme: "gemini", Host: "some.url"},
			expectedHost: "some.url:1965",
		},
		"with port": {
			url:          url.URL{Host: "some.url:2000"},
			expectedHost: "some.url:2000",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			host, err := buildHostString(tc.url)

			if tc.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error '%s', got nil", tc.expectedErr)
				}
				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Errorf("got error of '%s', want '%s'",
						err, tc.expectedErr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %q:", err)
				return
			}
			if host != tc.expectedHost{
				t.Errorf("got host string of '%s', want '%s'\n",
					host, tc.expectedHost)
			}
		})
	}
}

var (
	serverCert = "testdata/certs/server.crt"
	serverKey  = "testdata/certs/server.key"
)

// Server is a bare-bones test TCP server for responding to Gemini
// requests.
type Server struct {
	// Base URL of test server, no trailing slash
	URL string
	// Listener for requests
	listener net.Listener
}

// NewServer creates a new test server. Clients connect using
// a TLS dialer, so the test server needs to be configured with
// a TLS cert.
func NewServer() (*Server, error) {
	s := &Server{
		URL: "localhost:11965",
	}

	cert, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("could not load TLS certs: %w", err)
	}

	conf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	l, err := tls.Listen("tcp", ":11965", conf) // Shouldn't use default Gemini port, in case...
	if err != nil {
		return nil, fmt.Errorf("could not create test server: %w", err)
	}
	s.listener = l

	go s.run()

	return s, nil
}

// run processes connections from test clients.
func (s *Server) run() {
	// Match incoming URL to an associated test file
	responses := map[string]string{
		"gemini://localhost:11965/input":                          "./testdata/Input",
		"gemini://localhost:11965/redirect-temporary":             "./testdata/RedirectTemporary",
		"gemini://localhost:11965/redirect-permanent":             "./testdata/RedirectPermanent",
		"gemini://localhost:11965/cert-required":                  "./testdata/CertRequired",
		"gemini://localhost:11965/success":                        "./testdata/Success",
		"gemini://localhost:11965/failure-temporary":              "./testdata/FailureTemporary",
		"gemini://localhost:11965/failure-permanent":              "./testdata/FailurePermanent",
		"gemini://localhost:11965/invalid-header":                 "./testdata/InvalidHeader",
		"gemini://localhost:11965/redirected-temporarily-to-this": "./testdata/RedirectedTemporarilyToThis",
		"gemini://localhost:11965/redirected-permanently-to-this": "./testdata/RedirectedPermanentlyToThis",
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("could not accept connection: %s\n", err.Error())
			return
		}

		go func(conn net.Conn) {
			defer conn.Close()

			r := bufio.NewReader(conn)
			b, _, err := r.ReadLine()
			if err != nil {
				fmt.Printf("could not read request URL: %s\n", err)
				return
			}
			url := string(b)

			f, err := os.Open(responses[url])
			if err != nil {
				fmt.Printf("could not open test file '%s': %s\n", responses[url], err)
				return
			}
			defer f.Close()

			// Send header back, plus response body
			s := bufio.NewScanner(f)
			for s.Scan() {
				line := s.Text()
				conn.Write([]byte(line))
				conn.Write([]byte("\n"))
			}
		}(conn)
	}
}

func (s *Server) Close() {
	s.listener.Close()
}

func TestGet(t *testing.T) {
	svr, err := NewServer()
	if err != nil {
		t.Fatal("unable to start test server:", err)
	}
	defer svr.Close()

	testCases := map[string]struct {
		testURL        string
		testdata       string
		expectedStatus int
		expectedMeta   string
		expectedBody   string
		expectedErr    string
	}{
		"invalid url": {
			testURL:     "-gemini://",
			expectedErr: "error parsing supplied URL",
		},
		"invalid scheme": {
			testURL:     "http://localhost/",
			expectedErr: "unsupported URL scheme",
		},
		"invalid header": {
			testURL:     fmt.Sprintf("gemini://%s/invalid-header", svr.URL),
			expectedErr: "could not parse response header",
		},
		"input": {
			testURL:        fmt.Sprintf("gemini://%s/input", svr.URL),
			expectedMeta:   "Please enter a value",
			expectedStatus: StatusInput,
			expectedBody:   "",
			expectedErr:    "unsupported feature",
		},
		"temporary redirect": {
			testURL:        fmt.Sprintf("gemini://%s/redirect-temporary", svr.URL),
			expectedStatus: StatusSuccess,
			expectedMeta:   "text/gemini",
			expectedBody: `
# Should have been temporarily redirected here
		`,
		},
		"permanent redirect": {
			testURL:        fmt.Sprintf("gemini://%s/redirect-permanent", svr.URL),
			expectedStatus: StatusSuccess,
			expectedMeta:   "text/gemini",
			expectedBody: `
# Should have been permanently redirected here
		`,
		},
		"cert required": {
			testURL:        fmt.Sprintf("gemini://%s/cert-required", svr.URL),
			expectedStatus: StatusClientCertRequired,
			expectedMeta:   "Need a valid certificate",
			expectedBody:   "",
			expectedErr:    "resource requires a client certificate",
		},
		"success": {
			testURL:        fmt.Sprintf("gemini://%s/success", svr.URL),
			expectedStatus: StatusSuccess,
			expectedMeta:   "text/gemini",
			expectedBody: `# This is a top-level heading
Followed by some body text
* Bullet 1
* Bullet 2
`,
		},
		"success no scheme": {
			testURL:        fmt.Sprintf("%s/success", svr.URL),
			expectedStatus: StatusSuccess,
			expectedMeta:   "text/gemini",
			expectedBody: `# This is a top-level heading
Followed by some body text
* Bullet 1
* Bullet 2
`,
		},
		"failure temoporary": {
			testURL:        fmt.Sprintf("gemini://%s/failure-temporary", svr.URL),
			expectedStatus: StatusTemporaryFailure,
			expectedMeta:   "temporary failure",
			expectedBody:   "",
			expectedErr:    "error: temporary failure",
		},
		"failure permanent": {
			testURL:        fmt.Sprintf("gemini://%s/failure-permanent", svr.URL),
			expectedStatus: StatusPermanentFailure,
			expectedMeta:   "permanent failure",
			expectedBody:   "",
			expectedErr:    "error: permanent failure",
		},
	}

	client := DefaultClient
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			rsp, err := client.Get(tc.testURL)

			if tc.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error '%s', got nil", tc.expectedErr)
				}
				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Errorf("got error of '%s', want '%s'",
						err, tc.expectedErr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %q:", err)
				return
			}
			if rsp.StatusCode != tc.expectedStatus{
				t.Errorf("got status code of %d, want %d",
					rsp.StatusCode, tc.expectedStatus)
			}
			if rsp.Meta != tc.expectedMeta{
				t.Errorf("got meta of '%s', want '%s'",
					rsp.Meta, tc.expectedMeta)
			}
			if bytes.Equal(rsp.Body, []byte(tc.expectedBody)) {
				t.Errorf("got body of '%s', want '%s'",
					string(rsp.Body), tc.expectedBody)
			}
		})
	}
}
