package archive

import (
	"github.com/bhatti/api-mock-service/internal/types"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/url"
	"time"
)

/*
HTTP Archive (HAR) format
https://w3c.github.io/web-performance/specs/HAR/Overview.html
https://github.com/mrichman/hargo/blob/master/types.go
*/

// Har is a container type for deserialization
type Har struct {
	Log HarLog `json:"log"`
}

// HarLog represents the root of the exported data. This object MUST be present and its name MUST be "log".
type HarLog struct {
	// The object contains the following name/value pairs:

	// Required. Version number of the format.
	Version string `json:"version"`
	// Required. An object of type creator that contains the name and version
	// information of the log creator application.
	Creator HarCreator `json:"creator"`
	// Optional. An object of type browser that contains the name and version
	// information of the user agent.
	Browser HarBrowser `json:"browser"`
	// Optional. An array of objects of type page, each representing one exported
	// (tracked) page. Leave out this field if the application does not support
	// grouping by pages.
	Pages []HarPage `json:"pages,omitempty"`
	// Required. An array of objects of type entry, each representing one
	// exported (tracked) HTTP request.
	Entries []HarEntry `json:"entries"`
	// Optional. A comment provided by the user or the application. Sorting
	// entries by startedDateTime (starting from the oldest) is preferred way how
	// to export data since it can make importing faster. However, the reader
	// application should always make sure the array is sorted (if required for
	// the import).
	Comment string `json:"comment,omitempty"`
}

// HarCreator contains information about the log creator application
type HarCreator struct {
	// Required. The name of the application that created the log.
	Name string `json:"name"`
	// Required. The version number of the application that created the log.
	Version string `json:"version"`
	// Optional. A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// HarBrowser that created the log
type HarBrowser struct {
	// Required. The name of the browser that created the log.
	Name string `json:"name"`
	// Required. The version number of the browser that created the log.
	Version string `json:"version"`
	// Optional. A comment provided by the user or the browser.
	Comment string `json:"comment,omitempty"`
}

// HarPage object for every exported web page and one <entry> object for every HTTP request.
// In case when an HTTP trace tool isn't able to group requests by a page,
// the <pages> object is empty and individual requests doesn't have a parent page.
type HarPage struct {
	/* There is one <page> object for every exported web page and one <entry>
	object for every HTTP request. In case when an HTTP trace tool isn't able to
	group requests by a page, the <pages> object is empty and individual
	requests doesn't have a parent page.
	*/

	// Date and time stamp for the beginning of the page load
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.45+01:00).
	StartedDateTime string `json:"startedDateTime"`
	// Unique identifier of a page within the . Entries use it to refer the parent page.
	ID string `json:"id"`
	// HarPage title.
	Title string `json:"title"`
	// Detailed timing info about page load.
	PageTiming PageTiming `json:"pageTiming"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// PageTiming describes timings for various events (states) fired during the page load.
// All times are specified in milliseconds. If a time info is not available appropriate field is set to -1.
type PageTiming struct {
	// HarResponseContent of the page loaded. Number of milliseconds since page load started
	// (page.startedDateTime). Use -1 if the timing does not apply to the current
	// request.
	// Depending on the browser, onContentLoad property represents DOMContentLoad
	// event or document.readyState == interactive.
	OnContentLoad int `json:"onContentLoad"`
	// HarPage is loaded (onLoad event fired). Number of milliseconds since page
	// load started (page.startedDateTime). Use -1 if the timing does not apply
	// to the current request.
	OnLoad int `json:"onLoad"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// HarEntry is a unique, optional Reference to the parent page.
// Leave out this field if the application does not support grouping by pages.
type HarEntry struct {
	Title   string `json:"title,omitempty"`
	PageRef string `json:"pageref,omitempty"`
	// Date and time stamp of the request start
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD).
	StartedDateTime string `json:"startedDateTime"`
	// Total elapsed time of the request in milliseconds. This is the sum of all
	// timings available in the timings object (i.e. not including -1 values) .
	Time float32 `json:"time"`
	// Detailed info about the request.
	Request HarRequest `json:"request"`
	// Detailed info about the response.
	Response HarResponse `json:"response"`
	// Info about cache usage.
	Cache HarCache `json:"cache,omitempty"`
	// Detailed timing info about request/response round trip.
	PageTimings PageTimings `json:"pageTimings,omitempty"`
	// optional (new in 1.2) IP address of the server that was connected
	// (result of DNS resolution).
	ServerIPAddress string `json:"serverIPAddress,omitempty"`
	// optional (new in 1.2) Unique ID of the parent TCP/IP connection, can be
	// the client port number. Note that a port number doesn't have to be unique
	// identifier in cases where the port is shared for more connections. If the
	// port isn't available for the application, any other unique connection ID
	// can be used instead (e.g. connection index). Leave out this field if the
	// application doesn't support this info.
	Connection string `json:"connection,omitempty"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// HarRequest contains detailed info about performed request.
type HarRequest struct {
	// HarRequest method (GET, POST, ...).
	Method string `json:"method"`
	// Absolute URL of the request (fragments are not included).
	URL string `json:"url"`
	// HarRequest HTTP Version.
	HTTPVersion string `json:"httpVersion"`
	// List of cookie objects.
	Cookies []HarCookie `json:"cookies,omitempty"`
	// List of header objects.
	Headers []NVP `json:"headers,omitempty"`
	// List of query parameter objects.
	QueryString []NVP `json:"queryString,omitempty"`
	// Posted data.
	PostData HarPostData `json:"postData,omitempty"`
	// Total number of bytes from the start of the HTTP request message until
	// (and including) the double CRLF before the body. Set to -1 if the info
	// is not available.
	HeaderSize int `json:"headerSize"`
	// Size of the request body (POST data payload) in bytes. Set to -1 if the
	// info is not available.
	BodySize int `json:"bodySize"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// HarResponse contains detailed info about the response.
type HarResponse struct {
	// HarResponse status.
	Status int `json:"status"`
	// HarResponse status description.
	StatusText string `json:"statusText"`
	// HarResponse HTTP Version.
	HTTPVersion string `json:"httpVersion"`
	// List of cookie objects.
	Cookies []HarCookie `json:"cookies,omitempty"`
	// List of header objects.
	Headers []NVP `json:"headers,omitempty"`
	// Details about the response body.
	Content HarResponseContent `json:"content,omitempty"`
	// Redirection target URL from the Location response header.
	RedirectURL string `json:"redirectURL"`
	// Total number of bytes from the start of the HTTP response message until
	// (and including) the double CRLF before the body. Set to -1 if the info is
	// not available.
	// The size of received response-headers is computed only from headers that
	// are really received from the server. Additional headers appended by the
	// browser are not included in this number, but they appear in the list of
	// header objects.
	HeadersSize int `json:"headersSize"`
	// Size of the received response body in bytes. Set to zero in case of
	// responses coming from the cache (304). Set to -1 if the info is not
	// available.
	BodySize int `json:"bodySize"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// HarCookie contains list of all cookies (used in <request> and <response> objects).
type HarCookie struct {
	// The name of the cookie.
	Name string `json:"name"`
	// The cookie value.
	Value string `json:"value"`
	// optional The path pertaining to the cookie.
	Path string `json:"path,omitempty"`
	// optional The host of the cookie.
	Domain string `json:"domain,omitempty"`
	// optional HarCookie expiration time.
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.123+02:00).
	Expires string `json:"expires,omitempty"`
	// optional Set to true if the cookie is HTTP only, false otherwise.
	HTTPOnly bool `json:"httpOnly,omitempty"`
	// optional (new in 1.2) True if the cookie was transmitted over ssl, false
	// otherwise.
	Secure bool `json:"secure,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment bool `json:"comment,omitempty"`
}

// NVP is simply a name/value pair with a comment
type NVP struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

// HarPostData describes posted data, if any (embedded in <request> object).
type HarPostData struct {
	//  Mime type of posted data.
	MimeType string `json:"mimeType"`
	//  List of posted parameters (in case of URL encoded parameters).
	Params []HarPostParam `json:"params,omitempty"`
	//  Plain text posted data
	Text string `json:"text"`
	// optional (new in 1.2) A comment provided by the user or the
	// application.
	Comment string `json:"comment,omitempty"`
}

// HarPostParam is a list of posted parameters, if any (embedded in <postData> object).
type HarPostParam struct {
	// name of a posted parameter.
	Name string `json:"name"`
	// optional value of a posted parameter or content of a posted file.
	Value string `json:"value,omitempty"`
	// optional name of a posted file.
	FileName string `json:"fileName,omitempty"`
	// optional content type of posted file.
	ContentType string `json:"contentType,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// HarResponseContent describes details about response content (embedded in <response> object).
type HarResponseContent struct {
	// Length of the returned content in bytes. Should be equal to
	// response.bodySize if there is no compression and bigger when the content
	// has been compressed.
	Size int `json:"size"`
	// optional Number of bytes saved. Leave out this field if the information
	// is not available.
	Compression int `json:"compression,omitempty"`
	// MIME type of the response text (value of the HarResponseContent-Type response
	// header). The charset attribute of the MIME type is included (if
	// available).
	MimeType string `json:"mimeType"`
	// optional HarResponse body sent from the server or loaded from the browser
	// cache. This field is populated with textual content only. The text field
	// is either HTTP decoded text or encoded (e.g. "base64") representation of
	// the response body. Leave out this field if the information is not
	// available.
	Text string `json:"text,omitempty"`
	// optional (new in 1.2) Encoding used for response text field e.g
	// "base64". Leave out this field if the text field is HTTP decoded
	// (decompressed & unchunked), than trans-coded from its original character
	// set into UTF-8.
	Encoding string `json:"encoding,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
	// optional (community enhancement) A path to an attached file containing this content
	// used by Playwright
	File string `json:"_file,omitempty"`
}

// HarCache contains info about a request coming from browser cache.
type HarCache struct {
	// optional State of a cache entry before the request. Leave out this field
	// if the information is not available.
	BeforeRequest HarCacheObject `json:"beforeRequest,omitempty"`
	// optional State of a cache entry after the request. Leave out this field if
	// the information is not available.
	AfterRequest HarCacheObject `json:"afterRequest,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// HarCacheObject is used by both beforeRequest and afterRequest
type HarCacheObject struct {
	// optional - Expiration time of the cache entry.
	Expires string `json:"expires,omitempty"`
	// The last time the cache entry was opened.
	LastAccess string `json:"lastAccess"`
	// Etag
	ETag string `json:"eTag"`
	// The number of times the cache entry has been opened.
	HitCount int `json:"hitCount"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// PageTimings describes various phases within request-response round trip.
// All times are specified in milliseconds.
type PageTimings struct {
	Blocked int `json:"blocked,omitempty"`
	// optional - Time spent in a queue waiting for a network connection. Use -1
	// if the timing does not apply to the current request.
	DNS int `json:"dns,omitempty"`
	// optional - DNS resolution time. The time required to resolve a host name.
	// Use -1 if the timing does not apply to the current request.
	Connect int `json:"connect,omitempty"`
	// optional - Time required to create TCP connection. Use -1 if the timing
	// does not apply to the current request.
	Send int `json:"send"`
	// Time required to send HTTP request to the server.
	Wait int `json:"wait"`
	// Waiting for a response from the server.
	Receive int `json:"receive"`
	// Time required to read entire response from the server (or cache).
	Ssl int `json:"ssl,omitempty"`
	// optional (new in 1.2) - Time required for SSL/TLS negotiation. If this
	// field is defined then the time is also included in the connect field (to
	// ensure backward compatibility with HAR 1.1). Use -1 if the timing does not
	// apply to the current request.
	Comment string `json:"comment,omitempty"`
	// optional (new in 1.2) - A comment provided by the user or the application.
}

// BuildScenarios builds scenarios from HAR log
func BuildScenarios(
	config *types.Configuration,
	har *Har,
) (res []*types.APIScenario) {
	for _, entry := range har.Log.Entries {
		scenario, err := toScenario(config, entry)
		if err == nil {
			res = append(res, scenario)
		} else {
			log.WithFields(log.Fields{
				"entry": entry,
				"Error": err,
			}).Warnf("failed to import har file")
		}
	}
	return
}

// BuildHar extracts http request and builds HAR log
func BuildHar(
	config *types.Configuration,
	scenario *types.APIScenario,
	u *url.URL,
	started time.Time,
	ended time.Time) *Har {
	return &Har{
		Log: HarLog{
			Version: "1.2",
			Creator: HarCreator{
				Name:    config.UserAgent,
				Version: config.Version.String(),
				Comment: "",
			},
			Browser: HarBrowser{
				Name:    config.UserAgent,
				Version: config.Version.String(),
				Comment: "",
			},
			Pages: []HarPage{
				{
					StartedDateTime: started.UTC().Format(time.RFC3339),
					ID:              scenario.Group,
					Title:           scenario.Name,
					PageTiming: PageTiming{
						OnContentLoad: -1,
						OnLoad:        -1,
						Comment:       "",
					},
					Comment: "",
				}},
			Comment: "",
			Entries: []HarEntry{toEntry(scenario, u, started, ended)},
		},
	}
}

func toScenario(config *types.Configuration, entry HarEntry) (*types.APIScenario, error) {
	u, err := url.Parse(entry.Request.URL)
	if err != nil {
		return nil, err
	}
	scenario := types.BuildScenarioFromHTTP(
		config,
		"recorded-",
		u,
		entry.Request.Method,
		entry.PageRef,
		entry.Request.HTTPVersion,
		entry.Response.HTTPVersion,
		[]byte(entry.Request.PostData.Text),
		[]byte(entry.Response.Content.Text),
		nvpToMap(entry.Request.QueryString),
		postParamsToMap(entry.Request.PostData.Params),
		nvpToMap(entry.Request.Headers),
		entry.Request.PostData.MimeType,
		nvpToMap(entry.Response.Headers),
		entry.Response.Content.MimeType,
		entry.Response.Status)
	scenario.URL = entry.Request.URL
	scenario.StartTime, _ = time.Parse(entry.StartedDateTime, time.RFC3339)
	scenario.EndTime = scenario.StartTime.Add(time.Duration(entry.Time) * time.Millisecond)
	return scenario, nil
}

func toEntry(
	scenario *types.APIScenario,
	u *url.URL,
	started time.Time,
	ended time.Time) HarEntry {
	host, port, _ := net.SplitHostPort(u.Host)
	return HarEntry{
		Title:           scenario.Name,
		PageRef:         scenario.Group,
		StartedDateTime: started.UTC().Format(time.RFC3339),
		Time:            float32(ended.UnixMilli() - started.UnixMilli()),
		Request:         toRequest(scenario, u),
		Response:        toResponse(scenario),
		Cache:           HarCache{},
		PageTimings: PageTimings{
			Blocked: 0,
			DNS:     0,
			Connect: 0,
			Send:    0,
			Wait:    0,
			Receive: 0,
			Ssl:     0,
			Comment: "",
		},
		ServerIPAddress: host,
		Connection:      port,
		Comment:         "",
	}
}

func toRequest(
	scenario *types.APIScenario,
	u *url.URL,
) HarRequest {
	headers := toNVP(scenario.Request.Headers)
	headersSize := 0
	for _, header := range headers {
		headersSize += len(header.Name) + len(header.Value)
	}
	postData := HarPostData{
		MimeType: scenario.Request.ContentType(""),
		Params:   toPostParams(scenario),
		Text:     scenario.Request.Contents,
		Comment:  "",
	}
	return HarRequest{
		Method:      string(scenario.Method),
		URL:         u.Scheme + "://" + u.Host + scenario.Path,
		HTTPVersion: scenario.Request.HTTPVersion,
		Cookies:     toCookies(nil),
		Headers:     headers,
		QueryString: toNVP(scenario.Request.QueryParams),
		PostData:    postData,
		HeaderSize:  headersSize,
		BodySize:    len(scenario.Request.Contents),
		Comment:     "",
	}
}

func toResponse(scenario *types.APIScenario) HarResponse {
	headers := toNVPArray(scenario.Response.Headers)
	headersSize := 0
	for _, header := range headers {
		headersSize += len(header.Name) + len(header.Value)
	}
	var redirectURL string
	if len(scenario.Response.Headers["Location"]) > 0 {
		redirectURL = scenario.Response.Headers["Location"][0]
	}
	return HarResponse{
		Status:      scenario.Response.StatusCode,
		StatusText:  "",
		HTTPVersion: scenario.Response.HTTPVersion,
		Cookies:     toCookies(nil),
		Headers:     headers,
		Content: HarResponseContent{
			Size:        len(scenario.Response.Contents),
			Compression: 0,
			MimeType:    scenario.Response.ContentType(""),
			Text:        scenario.Response.Contents,
			Encoding:    "", // base64
			Comment:     "",
			File:        "",
		},
		RedirectURL: redirectURL,
		HeadersSize: headersSize,
		BodySize:    len(scenario.Response.Contents),
		Comment:     "",
	}
}

func toCookies(cookies []*http.Cookie) (res []HarCookie) {
	for _, cookie := range cookies {
		res = append(res, HarCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Expires:  cookie.Expires.UTC().Format(time.RFC3339),
			HTTPOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
		})
	}
	return
}

func toPostParams(scenario *types.APIScenario) (res []HarPostParam) {
	for k, v := range scenario.Request.PostParams {
		res = append(res, HarPostParam{
			// name of a posted parameter.
			Name:        k,
			Value:       v,
			FileName:    "",
			ContentType: "",
		})
	}
	return
}

func toNVP(store map[string]string) (res []NVP) {
	for k, v := range store {
		res = append(res, NVP{
			Name:    k,
			Value:   v,
			Comment: "",
		})
	}
	return
}

func toNVPArray(store map[string][]string) (res []NVP) {
	for k, vals := range store {
		for _, v := range vals {
			res = append(res, NVP{
				Name:    k,
				Value:   v,
				Comment: "",
			})
		}
	}
	return
}

func nvpToMap(nvpList []NVP) (res map[string][]string) {
	res = make(map[string][]string)
	for _, nvp := range nvpList {
		res[nvp.Name] = []string{nvp.Value}
	}
	return
}

func postParamsToMap(postParams []HarPostParam) (res map[string][]string) {
	res = make(map[string][]string)
	for _, nvp := range postParams {
		res[nvp.Name] = []string{nvp.Value}
	}
	return
}
