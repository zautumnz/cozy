package evaluator

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/zacanger/cozy/object"
)

// Code based on github.com/kirinlabs/HttpRequest, apache 2.0 licensed

type Request struct {
	cli               *http.Client
	transport         *http.Transport
	url               string
	method            string
	time              int64
	timeout           time.Duration
	headers           map[string]string
	username          string
	password          string
	data              interface{}
	disableKeepAlives bool
	tlsClientConfig   *tls.Config
	proxy             func(*http.Request) (*url.URL, error)
	checkRedirect     func(req *http.Request, via []*http.Request) error
}

func (r *Request) DisableKeepAlives(v bool) *Request {
	r.disableKeepAlives = v
	return r
}

func (r *Request) CheckRedirect(v func(req *http.Request, via []*http.Request) error) *Request {
	r.checkRedirect = v
	return r
}

// Build client
func (r *Request) buildClient() *http.Client {
	if r.cli == nil {
		r.cli = &http.Client{
			Transport:     http.DefaultTransport,
			CheckRedirect: r.checkRedirect,
			Timeout:       time.Second * r.timeout,
		}
	}
	return r.cli
}

// Set headers
func (r *Request) SetHeaders(headers map[string]string) *Request {
	if headers != nil || len(headers) > 0 {
		for k, v := range headers {
			r.headers[k] = v
		}
	}
	return r
}

// Init headers
func (r *Request) initHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "text/plain")
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
}

// Check application/json
func (r *Request) isJson() bool {
	if len(r.headers) > 0 {
		for _, v := range r.headers {
			if strings.Contains(strings.ToLower(v), "application/json") {
				return true
			}
		}
	}
	return false
}

func (r *Request) JSON() *Request {
	r.SetHeaders(map[string]string{"Content-Type": "application/json"})
	return r
}

// Build query data
func (r *Request) buildBody(d ...interface{}) (io.Reader, error) {
	if r.method == "GET" || r.method == "DELETE" || len(d) == 0 || (len(d) > 0 && d[0] == nil) {
		return nil, nil
	}

	switch d[0].(type) {
	case string:
		return strings.NewReader(d[0].(string)), nil
	case []byte:
		return bytes.NewReader(d[0].([]byte)), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return bytes.NewReader(IntByte(d[0])), nil
	case *bytes.Reader:
		return d[0].(*bytes.Reader), nil
	case *strings.Reader:
		return d[0].(*strings.Reader), nil
	case *bytes.Buffer:
		return d[0].(*bytes.Buffer), nil
	default:
		if r.isJson() {
			b, err := json.Marshal(d[0])
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(b), nil
		}
	}

	t := reflect.TypeOf(d[0]).String()
	if !strings.Contains(t, "map[string]interface") {
		return nil, errors.New("Unsupported data type.")
	}

	data := make([]string, 0)
	for k, v := range d[0].(map[string]interface{}) {
		if s, ok := v.(string); ok {
			data = append(data, fmt.Sprintf("%s=%v", k, s))
			continue
		}
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		data = append(data, fmt.Sprintf("%s=%s", k, string(b)))
	}

	return strings.NewReader(strings.Join(data, "&")), nil
}

func (r *Request) SetTimeout(d time.Duration) *Request {
	r.timeout = d
	return r
}

// Parse query for GET request
func parseQuery(url string) ([]string, error) {
	urlList := strings.Split(url, "?")
	if len(urlList) < 2 {
		return make([]string, 0), nil
	}
	query := make([]string, 0)
	for _, val := range strings.Split(urlList[1], "&") {
		v := strings.Split(val, "=")
		if len(v) < 2 {
			return make([]string, 0), errors.New("query parameter error")
		}
		query = append(query, fmt.Sprintf("%s=%s", v[0], v[1]))
	}
	return query, nil
}

// Build GET request url
func buildUrl(url string, data ...interface{}) (string, error) {
	query, err := parseQuery(url)
	if err != nil {
		return url, err
	}

	if len(data) > 0 && data[0] != nil {
		t := reflect.TypeOf(data[0]).String()
		switch t {
		case "map[string]interface {}":
			for k, v := range data[0].(map[string]interface{}) {
				vv := ""
				if reflect.TypeOf(v).String() == "string" {
					vv = v.(string)
				} else {
					b, err := json.Marshal(v)
					if err != nil {
						return url, err
					}
					vv = string(b)
				}
				query = append(query, fmt.Sprintf("%s=%s", k, vv))
			}
		case "string":
			param := data[0].(string)
			if param != "" {
				query = append(query, param)
			}
		default:
			return url, errors.New("Unsupported data type.")
		}

	}

	list := strings.Split(url, "?")

	if len(query) > 0 {
		return fmt.Sprintf("%s?%s", list[0], strings.Join(query, "&")), nil
	}

	return list[0], nil
}

func (r *Request) elapsedTime(n int64, resp *Response) {
	end := time.Now().UnixNano() / 1e6
	resp.time = end - n
}

// Get is a get http request
func (r *Request) Get(url string, data ...interface{}) (*Response, error) {
	return r.request(http.MethodGet, url, data...)
}

// Post is a post http request
func (r *Request) Post(url string, data ...interface{}) (*Response, error) {
	return r.request(http.MethodPost, url, data...)
}

// Put is a put http request
func (r *Request) Put(url string, data ...interface{}) (*Response, error) {
	return r.request(http.MethodPut, url, data...)
}

// Delete is a delete http request
func (r *Request) Delete(url string, data ...interface{}) (*Response, error) {
	return r.request(http.MethodDelete, url, data...)
}

// Upload file
func (r *Request) Upload(url, filename, fileinput string) (*Response, error) {
	return r.sendFile(url, filename, fileinput)
}

// Send http request
func (r *Request) request(method, url string, data ...interface{}) (*Response, error) {
	// Build Response
	response := &Response{}

	// Start time
	start := time.Now().UnixNano() / 1e6
	// Count elapsed time
	defer r.elapsedTime(start, response)

	if method == "" || url == "" {
		return nil, errors.New("parameter method and url is required")
	}

	r.url = url
	if len(data) > 0 {
		r.data = data[0]
	} else {
		r.data = ""
	}

	var (
		err  error
		req  *http.Request
		body io.Reader
	)
	r.cli = r.buildClient()

	method = strings.ToUpper(method)
	r.method = method

	if method == "GET" || method == "DELETE" {
		url, err = buildUrl(url, data...)
		if err != nil {
			return nil, err
		}
		r.url = url
	}

	body, err = r.buildBody(data...)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	r.initHeaders(req)

	resp, err := r.cli.Do(req)
	if err != nil {
		return nil, err
	}

	response.url = url
	response.resp = resp

	return response, nil
}

// Send file
func (r *Request) sendFile(url, filename, fileinput string) (*Response, error) {
	if url == "" {
		return nil, errors.New("parameter url is required")
	}

	fileBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(fileBuffer)
	fileWriter, er := bodyWriter.CreateFormFile(fileinput, filename)
	if er != nil {
		return nil, er
	}

	f, er := os.Open(filename)
	if er != nil {
		return nil, er
	}
	defer f.Close()

	_, er = io.Copy(fileWriter, f)
	if er != nil {
		return nil, er
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	// Build Response
	response := &Response{}

	// Start time
	start := time.Now().UnixNano() / 1e6
	// Count elapsed time
	defer r.elapsedTime(start, response)

	r.url = url
	r.data = nil

	var (
		err error
		req *http.Request
	)
	r.cli = r.buildClient()
	r.method = "POST"

	req, err = http.NewRequest(r.method, url, fileBuffer)
	if err != nil {
		return nil, err
	}

	r.initHeaders(req)
	req.Header.Set("Content-Type", contentType)

	resp, err := r.cli.Do(req)
	if err != nil {
		return nil, err
	}

	response.url = url
	response.resp = resp

	return response, nil
}

type Response struct {
	time int64
	url  string
	resp *http.Response
	body []byte
}

func (r *Response) Response() *http.Response {
	if r != nil {
		return r.resp
	}
	return nil
}

func (r *Response) StatusCode() int {
	if r.resp == nil {
		return 0
	}
	return r.resp.StatusCode
}

func (r *Response) Time() string {
	if r != nil {
		return fmt.Sprintf("%dms", r.time)
	}
	return "0ms"
}

func (r *Response) Url() string {
	if r != nil {
		return r.url
	}
	return ""
}

func (r *Response) Headers() http.Header {
	if r != nil {
		return r.resp.Header
	}
	return nil
}

func (r *Response) Body() ([]byte, error) {
	if r == nil {
		return []byte{}, errors.New("HttpRequest.Response is nil.")
	}

	defer r.resp.Body.Close()

	if len(r.body) > 0 {
		return r.body, nil
	}

	if r.resp == nil || r.resp.Body == nil {
		return nil, errors.New("response or body is nil")
	}

	b, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		return nil, err
	}
	r.body = b

	return b, nil
}

func (r *Response) Content() (string, error) {
	b, err := r.Body()
	if err != nil {
		return "", nil
	}
	return string(b), nil
}

func (r *Response) Json(v interface{}) error {
	return r.Unmarshal(v)
}

func (r *Response) Unmarshal(v interface{}) error {
	b, err := r.Body()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	return nil
}

func (r *Response) Close() error {
	if r != nil {
		return r.resp.Body.Close()
	}
	return nil
}

func Json(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func IntByte(v interface{}) []byte {
	b := bytes.NewBuffer([]byte{})
	switch v.(type) {
	case int:
		binary.Write(b, binary.BigEndian, int64(v.(int)))
	case int8:
		binary.Write(b, binary.BigEndian, v.(int8))
	case int16:
		binary.Write(b, binary.BigEndian, v.(int16))
	case int32:
		binary.Write(b, binary.BigEndian, v.(int32))
	case int64:
		binary.Write(b, binary.BigEndian, v.(int64))
	case uint:
		binary.Write(b, binary.BigEndian, uint64(v.(uint)))
	case uint8:
		binary.Write(b, binary.BigEndian, v.(uint8))
	case uint16:
		binary.Write(b, binary.BigEndian, v.(uint16))
	case uint32:
		binary.Write(b, binary.BigEndian, v.(uint32))
	case uint64:
		binary.Write(b, binary.BigEndian, v.(uint64))
	}
	return b.Bytes()
}

func newRequest() *Request {
	r := &Request{
		timeout: 60,
		headers: map[string]string{},
	}
	return r
}

func DisableKeepAlives(v bool) *Request {
	r := newRequest()
	return r.DisableKeepAlives(v)
}

func CheckRedirect(v func(req *http.Request, via []*http.Request) error) *Request {
	r := newRequest()
	return r.CheckRedirect(v)
}

func SetHeaders(headers map[string]string) *Request {
	r := newRequest()
	return r.SetHeaders(headers)
}

func JSON() *Request {
	r := newRequest()
	return r.JSON()
}

func SetTimeout(d time.Duration) *Request {
	r := newRequest()
	return r.SetTimeout(d)
}

// Get is a get http request
func Get(url string, data ...interface{}) (*Response, error) {
	r := newRequest()
	return r.Get(url, data...)
}

func Post(url string, data ...interface{}) (*Response, error) {
	r := newRequest()
	return r.Post(url, data...)
}

// Put is a put http request
func Put(url string, data ...interface{}) (*Response, error) {
	r := newRequest()
	return r.Put(url, data...)
}

// Delete is a delete http request
func Delete(url string, data ...interface{}) (*Response, error) {
	r := newRequest()
	return r.Delete(url, data...)
}

// Upload file
func Upload(url, filename, fileinput string) (*Response, error) {
	r := newRequest()
	return r.Upload(url, filename, fileinput)
}

func httpClient(args ...object.Object) object.Object {
	/*
		var uri string
		var method string
		var headers map[string]string
		var body string

		switch a := args[0].(type) {
		case *object.String:
			method = a.Value
		default:
			return NewError("http client expected method as first arg!")
		}
		switch a := args[1].(type) {
		case *object.String:
			uri = a.Value
		default:
			return NewError("http client expected uri as second arg!")
		}

		if len(args) > 2 {
			switch a := args[2].(type) {
			case *object.Hash:
				for _, pair := range a.Pairs {
					headers[pair.Key.Inspect()] = pair.Value.Inspect()
				}
			case *object.String:
				body = a.Value
			default:
				return NewError("http client expected headers or body as third arg!")
			}
		}

		if len(args) > 3 {
			switch a := args[3].(type) {
			case *object.String:
				body = a.Value
			default:
				return NewError("http client expected body as fourth arg!")
			}
		}
	*/

	ret := make(map[object.HashKey]object.HashPair)
	// TODO: switch on method, pass args, pull values out, put in ret

	resStatusKey := &object.String{Value: "status_code"}
	resStatusVal := &object.Integer{Value: 200} // TODO
	ret[resStatusKey.HashKey()] = object.HashPair{Key: resStatusKey, Value: resStatusVal}

	resProtoKey := &object.String{Value: "protocol"}
	resProtoVal := &object.String{Value: "HTTP/1.0"} // TODO
	ret[resProtoKey.HashKey()] = object.HashPair{Key: resProtoKey, Value: resProtoVal}

	resContentLengthKey := &object.String{Value: "content_length"}
	resContentLengthVal := &object.Integer{Value: 1} // TODO
	ret[resContentLengthKey.HashKey()] = object.HashPair{Key: resContentLengthKey, Value: resContentLengthVal}

	resBodyKey := &object.String{Value: "body"}
	resBodyVal := &object.Integer{Value: 1} // TODO
	ret[resBodyKey.HashKey()] = object.HashPair{Key: resBodyKey, Value: resBodyVal}

	// TODO: get headers out of map into this structure in another loop
	resHeaders := make(map[object.HashKey]object.HashPair)
	resHeadersKey := &object.String{Value: "headers"}
	resHeadersVal := &object.Hash{Pairs: resHeaders} // TODO
	ret[resHeadersKey.HashKey()] = object.HashPair{Key: resHeadersKey, Value: resHeadersVal}

	return &object.Hash{Pairs: ret}

}

func init() {
	RegisterBuiltin("http.client",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (httpClient(args...))
		})
}
