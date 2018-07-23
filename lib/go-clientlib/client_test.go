package clientlib

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/licensing/lib/errors"
	"github.com/stretchr/testify/require"
)

func TestJSONSuccessPath(t *testing.T) {
	t.Parallel()

	type sendBody struct {
		Sendfield string `json:"sendfield"`
	}

	type recvBody struct {
		Recvfield string `json:"recvfield"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NotZero(t, r.ContentLength)
		bits, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var send sendBody
		err = json.Unmarshal(bits, &send)
		require.NoError(t, err)
		require.Equal(t, "send1", send.Sendfield)

		bits, err = json.Marshal(recvBody{
			Recvfield: "recv1",
		})
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		w.Write(bits)
	}))
	defer server.Close()

	ctx := context.Background()

	sends := sendBody{Sendfield: "send1"}
	var recvs recvBody
	req, err := New(ctx, "GET", server.URL, SendJSON(sends), RecvJSON(&recvs))
	require.NoError(t, err)
	require.Equal(t, req.Method, "GET")
	require.NotEqual(t, req.ResponseHandle, DefaultResponseHandle)
	h := req.Header.Get("Accept")
	require.Equal(t, "application/json", h)
	h = req.Header.Get("Accept-Charset")
	require.Equal(t, "utf-8", h)
	h = req.Header.Get("Content-Type")
	require.Equal(t, "application/json", h)

	res, err := req.Do()
	require.NoError(t, err)
	h = res.Header.Get("Content-Type")
	require.Equal(t, "application/json", h)
	require.Equal(t, "recv1", recvs.Recvfield)

	// Body should have been closed
	_, err = ioutil.ReadAll(res.Body)
	require.Error(t, err)
	require.Regexp(t, "read on closed response body", err.Error())
}

func TestXMLSuccessPath(t *testing.T) {
	t.Parallel()

	type sendBody struct {
		Sendfield string `xml:"sendfield"`
	}

	type recvBody struct {
		Recvfield string `xml:"recvfield"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NotZero(t, r.ContentLength)
		bits, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var send sendBody
		err = xml.Unmarshal(bits, &send)
		require.NoError(t, err)
		require.Equal(t, "send1", send.Sendfield)

		bits, err = xml.Marshal(recvBody{
			Recvfield: "recv1",
		})
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/xml")
		w.Write(bits)
	}))
	defer server.Close()

	ctx := context.Background()

	sends := sendBody{Sendfield: "send1"}
	var recvs recvBody
	req, err := New(ctx, "GET", server.URL, SendXML(sends), RecvXML(&recvs))
	require.NoError(t, err)
	require.Equal(t, req.Method, "GET")
	require.NotEqual(t, req.ResponseHandle, DefaultResponseHandle)
	h := req.Header.Get("Accept")
	require.Equal(t, "application/xml", h)
	h = req.Header.Get("Accept-Charset")
	require.Equal(t, "utf-8", h)
	h = req.Header.Get("Content-Type")
	require.Equal(t, "application/xml", h)

	res, err := req.Do()
	require.NoError(t, err)
	h = res.Header.Get("Content-Type")
	require.Equal(t, "application/xml", h)
	require.Equal(t, "recv1", recvs.Recvfield)

	// Body should have been closed
	_, err = ioutil.ReadAll(res.Body)
	require.Error(t, err)
	require.Regexp(t, "read on closed response body", err.Error())
}

func TestTextSuccessPath(t *testing.T) {
	t.Parallel()

	sends := "input"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NotZero(t, r.ContentLength)
		bits, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		require.Equal(t, sends, string(bits))

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("output"))
	}))
	defer server.Close()

	ctx := context.Background()

	var recvs string
	req, err := New(ctx, "GET", server.URL, SendText(sends), RecvText(&recvs))
	require.NoError(t, err)
	require.Equal(t, req.Method, "GET")
	require.NotEqual(t, req.ResponseHandle, DefaultResponseHandle)
	h := req.Header.Get("Accept")
	require.Equal(t, "text/plain", h)
	h = req.Header.Get("Accept-Charset")
	require.Equal(t, "utf-8", h)
	h = req.Header.Get("Content-Type")
	require.Equal(t, "text/plain", h)

	res, err := req.Do()
	require.NoError(t, err)
	h = res.Header.Get("Content-Type")
	require.Equal(t, "text/plain", h)
	require.Equal(t, "output", recvs)

	// Body should have been closed
	_, err = ioutil.ReadAll(res.Body)
	require.Error(t, err)
	require.Regexp(t, "read on closed response body", err.Error())
}

func TestDefaultErrorChecker(t *testing.T) {
	t.Parallel()

	testStatus := map[int]bool{
		200: true,
		300: false,
		400: false,
		500: false,
	}

	tfunc := func(status int, good bool) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))
		defer server.Close()

		req, res, err := Do(context.Background(), "GET", server.URL)
		require.NotNil(t, req)
		require.NotNil(t, res)
		require.Equal(t, res.StatusCode, status)
		if good {
			require.NoError(t, err)
			require.Equal(t, status, res.StatusCode)
		} else {
			require.Error(t, err)
			s, ok := errors.HTTPStatus(err)
			require.True(t, ok)
			require.Equal(t, status, s)
		}
	}

	for status, good := range testStatus {
		tfunc(status, good)
	}
}

func TestErrorBodyMaxLength(t *testing.T) {
	t.Parallel()

	const errBodyMaxLength = 1337

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := "teststring"
		for i := 0; i < 10; i++ {
			s = s + s
		}
		b := []byte(s)
		require.True(t, len(b) > errBodyMaxLength)
		w.WriteHeader(599)
		w.Write(b)
	}))
	defer server.Close()

	errorCheckerCalled := false
	req, err := New(context.Background(), "GET", server.URL)
	require.NoError(t, err)

	req.ErrorBodyMaxLength = errBodyMaxLength

	req.ErrorCheck = func(r *Request, doErr error, res *http.Response) error {
		require.Equal(t, 599, res.StatusCode)
		errorCheckerCalled = true
		return DefaultErrorCheck(r, doErr, res)
	}

	errorSummaryCalled := false
	req.ErrorSummary = func(body []byte) string {
		require.True(t, len(body) <= errBodyMaxLength)
		errorSummaryCalled = true
		return DefaultErrorSummary(body)
	}

	res, err := req.Do()
	require.Error(t, err)
	require.NotNil(t, res)
	require.Equal(t, 599, res.StatusCode)
	require.True(t, errorCheckerCalled)
	require.True(t, errorSummaryCalled)
}

func TestBodyClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("teststring"))
	}))
	defer server.Close()

	_, res, err := Do(context.Background(), "GET", server.URL)
	require.NoError(t, err)
	// Body should have been closed
	_, err = ioutil.ReadAll(res.Body)
	require.Error(t, err)
	require.Regexp(t, "read on closed response body", err.Error())

	_, res, err = Do(context.Background(), "GET", server.URL, DontClose())
	require.NoError(t, err)
	// Body should not have been closed
	bits, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "teststring", string(bits))
	err = res.Body.Close()
	require.NoError(t, err)
}

func TestErrorConfigure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := "teststring"
		for i := 0; i < 5; i++ {
			s = s + s
		}
		b := []byte(s)
		require.True(t, len(b) > defaultErrBodyMaxLength)
		w.WriteHeader(599)
		w.Write(b)
	}))
	defer server.Close()

	errorCheckerCalled := false
	req, err := New(context.Background(), "GET", server.URL)
	require.NoError(t, err)

	req.ErrorCheck = func(r *Request, doErr error, res *http.Response) error {
		require.Equal(t, 599, res.StatusCode)
		errorCheckerCalled = true
		return DefaultErrorCheck(r, doErr, res)
	}

	errorSummaryCalled := false
	req.ErrorSummary = func(body []byte) string {
		require.True(t, len(body) <= defaultErrBodyMaxLength)
		errorSummaryCalled = true
		return DefaultErrorSummary(body)
	}

	res, err := req.Do()
	require.Error(t, err)
	require.NotNil(t, res)
	require.Equal(t, 599, res.StatusCode)
	require.True(t, errorCheckerCalled)
	require.True(t, errorSummaryCalled)
}
