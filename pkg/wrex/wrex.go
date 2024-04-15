package wrex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// The Opts struct is used to pass options to any Wrex request method.
type Opts struct {
	Url     string
	Headers map[string]string
	Body    interface{}
	Params  map[string]string

	// ValidateStatus is an optional function that can be used to validate the status code of the response.
	// If the function returns true, the response is considered valid, else it is considered invalid and an error is returned.
	//
	// The default implementation is `statusCode >= 200 && statusCode < 300`.
	ValidateStatus func(statusCode int) bool
}

type Resp struct {
	Data []byte
	*http.Response
}

func (r *Resp) String() string {
	return string(r.Data)
}

func (r *Resp) StatusCode() int {
	if r.Response == nil {
		return 0
	}
	return r.Response.StatusCode
}

type Client struct {
	client  *http.Client
	BaseUrl string
	Opts    Opts
}

// NewClient creates a new Wrex client with the given base URL.
func NewClient(baseUrl string, defaultOpts Opts) *Client {
	client := http.DefaultClient
	return &Client{client, baseUrl, defaultOpts}
}

// Get sends a GET request to the given URL.
func (c *Client) Get(ctx context.Context, opts Opts) (*Resp, error) {
	return c.Request(ctx, "GET", opts)
}

// Post sends a POST request to the given URL.
func (c *Client) Post(ctx context.Context, opts Opts) (*Resp, error) {
	return c.Request(ctx, "POST", opts)
}

// Put sends a PUT request to the given URL.
func (c *Client) Put(ctx context.Context, opts Opts) (*Resp, error) {
	return c.Request(ctx, "PUT", opts)
}

// Delete sends a DELETE request to the given URL.
func (c *Client) Delete(ctx context.Context, opts Opts) (*Resp, error) {
	return c.Request(ctx, "DELETE", opts)
}

var defaultClient = NewClient("", Opts{})

func Get(ctx context.Context, opts Opts) (*Resp, error) {
	return defaultClient.Get(ctx, opts)
}

func Post(ctx context.Context, opts Opts) (*Resp, error) {
	return defaultClient.Post(ctx, opts)
}

func Put(ctx context.Context, opts Opts) (*Resp, error) {
	return defaultClient.Put(ctx, opts)
}

func Delete(ctx context.Context, opts Opts) (*Resp, error) {
	return defaultClient.Delete(ctx, opts)
}

func Request(ctx context.Context, method string, opts Opts) (*Resp, error) {
	return defaultClient.Request(ctx, method, opts)
}

// JSON methods only provide as direct methods because you can't do the generic type thing with interface methods.
func GetJson[T any](ctx context.Context, opts Opts, v interface{}) (*Resp, error) {
	return jsonRequest[T](defaultClient, ctx, "GET", opts, v)
}

func PostJson[T any](ctx context.Context, opts Opts, v interface{}) (*Resp, error) {
	return jsonRequest[T](defaultClient, ctx, "POST", opts, v)
}

func PutJson[T any](ctx context.Context, opts Opts, v interface{}) (*Resp, error) {
	return jsonRequest[T](defaultClient, ctx, "PUT", opts, v)
}

func DeleteJson[T any](ctx context.Context, opts Opts, v interface{}) (*Resp, error) {
	return jsonRequest[T](defaultClient, ctx, "DELETE", opts, v)
}

func jsonRequest[T any](
	c *Client,
	ctx context.Context,
	method string,
	opts Opts,
	dst interface{},
) (*Resp, error) {
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	if _, ok := opts.Headers["Accept"]; !ok {
		opts.Headers["Accept"] = "application/json"
	}
	if _, ok := opts.Headers["Content-Type"]; !ok {
		opts.Headers["Content-Type"] = "application/json"
	}

	var resp *Resp
	var err error
	switch method {
	case "GET":
		resp, err = c.Get(ctx, opts)
	case "POST":
		resp, err = c.Post(ctx, opts)
	case "PUT":
		resp, err = c.Put(ctx, opts)
	case "DELETE":
		resp, err = c.Delete(ctx, opts)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return resp, err
	}

	if dst != nil {
		err = json.Unmarshal(resp.Data, dst)
	}

	return resp, err
}

func DefaultValidateStatus(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// Request sends a request to the given URL with the given method.
func (c *Client) Request(ctx context.Context, method string, opts Opts) (*Resp, error) {
	var payload []byte
	var err error

	switch opts.Body.(type) {
	case url.Values:
		payload = []byte(opts.Body.(url.Values).Encode())
		if opts.Headers == nil {
			opts.Headers = make(map[string]string)
		}
		opts.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	default:
		if opts.Body == nil {
			break
		}
		payload, err = json.Marshal(opts.Body)
		if err != nil {
			return nil, err
		}
	}

	url := c.BaseUrl + opts.Url
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, err
	}

	if opts.Params != nil {
		q := req.URL.Query()
		for k, v := range opts.Params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range opts.Headers {
		req.Header.Add(k, v)
	}

	rawResp, err := c.client.Do(req)
	resp := &Resp{Response: rawResp}
	if err != nil {
		return resp, err
	}

	defer resp.Body.Close()
	resp.Data, err = io.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	validate := DefaultValidateStatus
	if opts.ValidateStatus != nil {
		validate = opts.ValidateStatus
	}

	if !validate(resp.StatusCode()) {
		return resp, fmt.Errorf("invalid status code: %d", resp.StatusCode())
	}

	return resp, nil
}
