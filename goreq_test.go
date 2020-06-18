package goreq

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	gob "github.com/franela/goblin"
	m "github.com/onsi/gomega"
)

type Query struct {
	Limit int
	Skip  int
}

func TestRequest(t *testing.T) {

	query := Query{
		Limit: 3,
		Skip:  5,
	}

	valuesQuery := url.Values{}
	valuesQuery.Set("name", "marcos")
	valuesQuery.Add("friend", "jonas")
	valuesQuery.Add("friend", "peter")

	g := gob.Goblin(t)

	m.RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Request", func() {

		g.Describe("General request methods", func() {
			var ts *httptest.Server
			var requestHeaders http.Header

			g.Before(func() {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestHeaders = r.Header
					if (r.Method == "GET" || r.Method == "OPTIONS" || r.Method == "TRACE" || r.Method == "PATCH" || r.Method == "FOOBAR") && r.URL.Path == "/foo" {
						w.WriteHeader(200)
						fmt.Fprint(w, "bar")
					}
					if r.Method == "GET" && r.URL.Path == "/getquery" {
						w.WriteHeader(200)
						fmt.Fprint(w, fmt.Sprintf("%v", r.URL))
					}
					if r.Method == "GET" && r.URL.Path == "/getbody" {
						w.WriteHeader(200)
						io.Copy(w, r.Body)
					}
					if r.Method == "POST" && r.URL.Path == "/" {
						w.Header().Add("Location", ts.URL+"/123")
						w.WriteHeader(201)
						io.Copy(w, r.Body)
					}
					if r.Method == "POST" && r.URL.Path == "/getquery" {
						w.WriteHeader(200)
						fmt.Fprint(w, fmt.Sprintf("%v", r.URL))
					}
					if r.Method == "PUT" && r.URL.Path == "/foo/123" {
						w.WriteHeader(200)
						io.Copy(w, r.Body)
					}
					if r.Method == "DELETE" && r.URL.Path == "/foo/123" {
						w.WriteHeader(204)
					}
					if r.Method == "GET" && r.URL.Path == "/redirect_test/301" {
						http.Redirect(w, r, "/redirect_test/302", 301)
					}
					if r.Method == "GET" && r.URL.Path == "/redirect_test/302" {
						http.Redirect(w, r, "/redirect_test/303", 302)
					}
					if r.Method == "GET" && r.URL.Path == "/redirect_test/303" {
						http.Redirect(w, r, "/redirect_test/307", 303)
					}
					if r.Method == "GET" && r.URL.Path == "/redirect_test/307" {
						http.Redirect(w, r, "/getquery", 307)
					}
					if r.Method == "GET" && r.URL.Path == "/redirect_test/destination" {
						http.Redirect(w, r, ts.URL+"/destination", 301)
					}
					if r.Method == "GET" && r.URL.Path == "/getcookies" {
						defer r.Body.Close()
						w.WriteHeader(200)
						fmt.Fprint(w, requestHeaders.Get("Cookie"))
					}
					if r.Method == "GET" && r.URL.Path == "/setcookies" {
						defer r.Body.Close()
						w.Header().Add("Set-Cookie", "foobar=42 ; Path=/")
						w.WriteHeader(200)
					}
					if r.Method == "GET" && r.URL.Path == "/compressed" {
						defer r.Body.Close()
						b := "{\"foo\":\"bar\",\"fuu\":\"baz\"}"
						gw := gzip.NewWriter(w)
						defer gw.Close()
						if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
							w.Header().Add("Content-Encoding", "gzip")
						}
						w.WriteHeader(200)
						gw.Write([]byte(b))
					}
					if r.Method == "GET" && r.URL.Path == "/compressed_deflate" {
						defer r.Body.Close()
						b := "{\"foo\":\"bar\",\"fuu\":\"baz\"}"
						gw := zlib.NewWriter(w)
						defer gw.Close()
						if strings.Contains(r.Header.Get("Content-Encoding"), "deflate") {
							w.Header().Add("Content-Encoding", "deflate")
						}
						w.WriteHeader(200)
						gw.Write([]byte(b))
					}
					if r.Method == "GET" && r.URL.Path == "/compressed_and_return_compressed_without_header" {
						defer r.Body.Close()
						b := "{\"foo\":\"bar\",\"fuu\":\"baz\"}"
						gw := gzip.NewWriter(w)
						defer gw.Close()
						w.WriteHeader(200)
						gw.Write([]byte(b))
					}
					if r.Method == "GET" && r.URL.Path == "/compressed_deflate_and_return_compressed_without_header" {
						defer r.Body.Close()
						b := "{\"foo\":\"bar\",\"fuu\":\"baz\"}"
						gw := zlib.NewWriter(w)
						defer gw.Close()
						w.WriteHeader(200)
						gw.Write([]byte(b))
					}
					if r.Method == "POST" && r.URL.Path == "/compressed" && r.Header.Get("Content-Encoding") == "gzip" {
						defer r.Body.Close()
						gr, _ := gzip.NewReader(r.Body)
						defer gr.Close()
						b, _ := ioutil.ReadAll(gr)
						w.WriteHeader(201)
						w.Write(b)
					}
					if r.Method == "POST" && r.URL.Path == "/compressed_deflate" && r.Header.Get("Content-Encoding") == "deflate" {
						defer r.Body.Close()
						gr, _ := zlib.NewReader(r.Body)
						defer gr.Close()
						b, _ := ioutil.ReadAll(gr)
						w.WriteHeader(201)
						w.Write(b)
					}
					if r.Method == "POST" && r.URL.Path == "/compressed_and_return_compressed" {
						defer r.Body.Close()
						w.Header().Add("Content-Encoding", "gzip")
						w.WriteHeader(201)
						io.Copy(w, r.Body)
					}
					if r.Method == "POST" && r.URL.Path == "/compressed_deflate_and_return_compressed" {
						defer r.Body.Close()
						w.Header().Add("Content-Encoding", "deflate")
						w.WriteHeader(201)
						io.Copy(w, r.Body)
					}
					if r.Method == "POST" && r.URL.Path == "/compressed_deflate_and_return_compressed_without_header" {
						defer r.Body.Close()
						w.WriteHeader(201)
						io.Copy(w, r.Body)
					}
					if r.Method == "POST" && r.URL.Path == "/compressed_and_return_compressed_without_header" {
						defer r.Body.Close()
						w.WriteHeader(201)
						io.Copy(w, r.Body)
					}
				}))
			})

			g.After(func() {
				ts.Close()
			})

			g.Describe("GET", func() {

				g.It("Should do a GET", func() {
					res, err := Request{Uri: ts.URL + "/foo"}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("bar"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Should return ContentLength", func() {
					res, err := Request{Uri: ts.URL + "/foo"}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("bar"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
					m.Expect(res.ContentLength).Should(m.Equal(int64(3)))
				})

				g.It("Should do a GET with querystring", func() {
					res, err := Request{
						Uri:         ts.URL + "/getquery",
						QueryString: query,
					}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("/getquery?limit=3&skip=5"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Should support url.Values in querystring", func() {
					res, err := Request{
						Uri:         ts.URL + "/getquery",
						QueryString: valuesQuery,
					}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("/getquery?friend=jonas&friend=peter&name=marcos"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Should support sending string body", func() {
					res, err := Request{Uri: ts.URL + "/getbody", Body: "foo"}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("foo"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Shoulds support sending a Reader body", func() {
					res, err := Request{Uri: ts.URL + "/getbody", Body: strings.NewReader("foo")}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("foo"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Support sending any object that is json encodable", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Uri: ts.URL + "/getbody", Body: obj}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Support sending an array of bytes body", func() {
					bdy := []byte{'f', 'o', 'o'}
					res, err := Request{Uri: ts.URL + "/getbody", Body: bdy}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("foo"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Should return an error when body is not JSON encodable", func() {
					res, err := Request{Uri: ts.URL + "/getbody", Body: math.NaN()}.Do()

					m.Expect(res).Should(m.BeNil())
					m.Expect(err).ShouldNot(m.BeNil())
				})

				g.It("Should return a gzip reader if Content-Encoding is 'gzip'", func() {
					res, err := Request{Uri: ts.URL + "/compressed", Compression: Gzip()}.Do()
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(err).Should(m.BeNil())
					m.Expect(res.Body.compressedReader).ShouldNot(m.BeNil())
					m.Expect(res.Body.reader).ShouldNot(m.BeNil())
					m.Expect(string(b)).Should(m.Equal("{\"foo\":\"bar\",\"fuu\":\"baz\"}"))
					m.Expect(res.Body.compressedReader).ShouldNot(m.BeNil())
					m.Expect(res.Body.reader).ShouldNot(m.BeNil())
				})

				g.It("Should close reader and compresserReader on Body close", func() {
					res, err := Request{Uri: ts.URL + "/compressed", Compression: Gzip()}.Do()
					m.Expect(err).Should(m.BeNil())

					_, e := ioutil.ReadAll(res.Body.reader)
					m.Expect(e).Should(m.BeNil())
					_, e = ioutil.ReadAll(res.Body.compressedReader)
					m.Expect(e).Should(m.BeNil())

					_, e = ioutil.ReadAll(res.Body.reader)
					//when reading body again it doesnt error
					m.Expect(e).Should(m.BeNil())

					res.Body.Close()
					_, e = ioutil.ReadAll(res.Body.reader)
					//error because body is already closed
					m.Expect(e).ShouldNot(m.BeNil())

					_, e = ioutil.ReadAll(res.Body.compressedReader)
					//compressedReaders dont error on reading when closed
					m.Expect(e).Should(m.BeNil())
				})

				g.It("Should not return a gzip reader if Content-Encoding is not 'gzip'", func() {
					res, err := Request{Uri: ts.URL + "/compressed_and_return_compressed_without_header", Compression: Gzip()}.Do()
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(err).Should(m.BeNil())
					m.Expect(string(b)).ShouldNot(m.Equal("{\"foo\":\"bar\",\"fuu\":\"baz\"}"))
				})

				g.It("Should return a deflate reader if Content-Encoding is 'deflate'", func() {
					res, err := Request{Uri: ts.URL + "/compressed_deflate", Compression: Deflate()}.Do()
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(err).Should(m.BeNil())
					m.Expect(string(b)).Should(m.Equal("{\"foo\":\"bar\",\"fuu\":\"baz\"}"))
				})

				g.It("Should not return a delfate reader if Content-Encoding is not 'deflate'", func() {
					res, err := Request{Uri: ts.URL + "/compressed_deflate_and_return_compressed_without_header", Compression: Deflate()}.Do()
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(err).Should(m.BeNil())
					m.Expect(string(b)).ShouldNot(m.Equal("{\"foo\":\"bar\",\"fuu\":\"baz\"}"))
				})

				g.It("Should return a deflate reader when using zlib if Content-Encoding is 'deflate'", func() {
					res, err := Request{Uri: ts.URL + "/compressed_deflate", Compression: Zlib()}.Do()
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(err).Should(m.BeNil())
					m.Expect(string(b)).Should(m.Equal("{\"foo\":\"bar\",\"fuu\":\"baz\"}"))
				})

				g.It("Should not return a delfate reader when using zlib if Content-Encoding is not 'deflate'", func() {
					res, err := Request{Uri: ts.URL + "/compressed_deflate_and_return_compressed_without_header", Compression: Zlib()}.Do()
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(err).Should(m.BeNil())
					m.Expect(string(b)).ShouldNot(m.Equal("{\"foo\":\"bar\",\"fuu\":\"baz\"}"))
				})

				g.It("Should send cookies from the cookiejar", func() {
					uri, err := url.Parse(ts.URL + "/getcookies")
					m.Expect(err).Should(m.BeNil())

					jar, err := cookiejar.New(nil)
					m.Expect(err).Should(m.BeNil())

					jar.SetCookies(uri, []*http.Cookie{
						{
							Name:  "bar",
							Value: "foo",
							Path:  "/",
						},
					})

					res, err := Request{
						Uri:       ts.URL + "/getcookies",
						CookieJar: jar,
					}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("bar=foo"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
					m.Expect(res.ContentLength).Should(m.Equal(int64(7)))
				})

				g.It("Should send cookies added with .AddCookie", func() {
					c1 := &http.Cookie{Name: "c1", Value: "v1"}
					c2 := &http.Cookie{Name: "c2", Value: "v2"}

					req := Request{Uri: ts.URL + "/getcookies"}
					req.AddCookie(c1)
					req.AddCookie(c2)

					res, err := req.Do()
					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("c1=v1; c2=v2"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
					m.Expect(res.ContentLength).Should(m.Equal(int64(12)))
				})

				g.It("Should send cookies added with .WithCookie", func() {
					c1 := &http.Cookie{Name: "c1", Value: "v2"}
					c2 := &http.Cookie{Name: "c2", Value: "v3"}

					res, err := Request{Uri: ts.URL + "/getcookies"}.
						WithCookie(c1).
						WithCookie(c2).
						Do()
					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("c1=v2; c2=v3"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
					m.Expect(res.ContentLength).Should(m.Equal(int64(12)))
				})

				g.It("Should populate the cookiejar", func() {
					uri, err := url.Parse(ts.URL + "/setcookies")
					m.Expect(err).Should(m.BeNil())

					jar, _ := cookiejar.New(nil)
					m.Expect(err).Should(m.BeNil())

					res, err := Request{
						Uri:       ts.URL + "/setcookies",
						CookieJar: jar,
					}.Do()

					m.Expect(err).Should(m.BeNil())

					m.Expect(res.Header.Get("Set-Cookie")).Should(m.Equal("foobar=42 ; Path=/"))

					cookies := jar.Cookies(uri)
					m.Expect(len(cookies)).Should(m.Equal(1))

					cookie := cookies[0]
					m.Expect(*cookie).Should(m.Equal(http.Cookie{
						Name:  "foobar",
						Value: "42",
					}))
				})
			})

			g.Describe("POST", func() {
				g.It("Should send a string", func() {
					res, err := Request{Method: "POST", Uri: ts.URL, Body: "foo"}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("foo"))
					m.Expect(res.StatusCode).Should(m.Equal(201))
					m.Expect(res.Header.Get("Location")).Should(m.Equal(ts.URL + "/123"))
				})

				g.It("Should send a Reader", func() {
					res, err := Request{Method: "POST", Uri: ts.URL, Body: strings.NewReader("foo")}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("foo"))
					m.Expect(res.StatusCode).Should(m.Equal(201))
					m.Expect(res.Header.Get("Location")).Should(m.Equal(ts.URL + "/123"))
				})

				g.It("Send any object that is json encodable", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL, Body: obj}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
					m.Expect(res.Header.Get("Location")).Should(m.Equal(ts.URL + "/123"))
				})

				g.It("Send an array of bytes", func() {
					bdy := []byte{'f', 'o', 'o'}
					res, err := Request{Method: "POST", Uri: ts.URL, Body: bdy}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("foo"))
					m.Expect(res.StatusCode).Should(m.Equal(201))
					m.Expect(res.Header.Get("Location")).Should(m.Equal(ts.URL + "/123"))
				})

				g.It("Should return an error when body is not JSON encodable", func() {
					res, err := Request{Method: "POST", Uri: ts.URL, Body: math.NaN()}.Do()

					m.Expect(res).Should(m.BeNil())
					m.Expect(err).ShouldNot(m.BeNil())
				})

				g.It("Should do a POST with querystring", func() {
					bdy := []byte{'f', 'o', 'o'}
					res, err := Request{
						Method:      "POST",
						Uri:         ts.URL + "/getquery",
						Body:        bdy,
						QueryString: query,
					}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("/getquery?limit=3&skip=5"))
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Should send body as gzip if compressed", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed", Body: obj, Compression: Gzip()}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as deflate if compressed", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_deflate", Body: obj, Compression: Deflate()}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as deflate using zlib if compressed", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_deflate", Body: obj, Compression: Zlib()}.Do()

					m.Expect(err).Should(m.BeNil())
					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as gzip if compressed and parse return body", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_and_return_compressed", Body: obj, Compression: Gzip()}.Do()

					m.Expect(err).Should(m.BeNil())
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(string(b)).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as deflate if compressed and parse return body", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_deflate_and_return_compressed", Body: obj, Compression: Deflate()}.Do()

					m.Expect(err).Should(m.BeNil())
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(string(b)).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as deflate using zlib if compressed and parse return body", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_deflate_and_return_compressed", Body: obj, Compression: Zlib()}.Do()

					m.Expect(err).Should(m.BeNil())
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(string(b)).Should(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as gzip if compressed and not parse return body if header not set ", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_and_return_compressed_without_header", Body: obj, Compression: Gzip()}.Do()

					m.Expect(err).Should(m.BeNil())
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(string(b)).ShouldNot(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as deflate if compressed and not parse return body if header not set ", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_deflate_and_return_compressed_without_header", Body: obj, Compression: Deflate()}.Do()

					m.Expect(err).Should(m.BeNil())
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(string(b)).ShouldNot(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})

				g.It("Should send body as deflate using zlib if compressed and not parse return body if header not set ", func() {
					obj := map[string]string{"foo": "bar"}
					res, err := Request{Method: "POST", Uri: ts.URL + "/compressed_deflate_and_return_compressed_without_header", Body: obj, Compression: Zlib()}.Do()

					m.Expect(err).Should(m.BeNil())
					b, _ := ioutil.ReadAll(res.Body)
					m.Expect(string(b)).ShouldNot(m.Equal(`{"foo":"bar"}`))
					m.Expect(res.StatusCode).Should(m.Equal(201))
				})
			})

			g.It("Should do a PUT", func() {
				res, err := Request{Method: "PUT", Uri: ts.URL + "/foo/123", Body: "foo"}.Do()

				m.Expect(err).Should(m.BeNil())
				str, _ := res.Body.ToString()
				m.Expect(str).Should(m.Equal("foo"))
				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.It("Should do a DELETE", func() {
				res, err := Request{Method: "DELETE", Uri: ts.URL + "/foo/123"}.Do()

				m.Expect(err).Should(m.BeNil())
				m.Expect(res.StatusCode).Should(m.Equal(204))
			})

			g.It("Should do a OPTIONS", func() {
				res, err := Request{Method: "OPTIONS", Uri: ts.URL + "/foo"}.Do()

				m.Expect(err).Should(m.BeNil())
				str, _ := res.Body.ToString()
				m.Expect(str).Should(m.Equal("bar"))
				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.It("Should do a PATCH", func() {
				res, err := Request{Method: "PATCH", Uri: ts.URL + "/foo"}.Do()

				m.Expect(err).Should(m.BeNil())
				str, _ := res.Body.ToString()
				m.Expect(str).Should(m.Equal("bar"))
				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.It("Should do a TRACE", func() {
				res, err := Request{Method: "TRACE", Uri: ts.URL + "/foo"}.Do()

				m.Expect(err).Should(m.BeNil())
				str, _ := res.Body.ToString()
				m.Expect(str).Should(m.Equal("bar"))
				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.It("Should do a custom method", func() {
				res, err := Request{Method: "FOOBAR", Uri: ts.URL + "/foo"}.Do()

				m.Expect(err).Should(m.BeNil())
				str, _ := res.Body.ToString()
				m.Expect(str).Should(m.Equal("bar"))
				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.Describe("Responses", func() {
				g.It("Should handle strings", func() {
					res, _ := Request{Method: "POST", Uri: ts.URL, Body: "foo bar"}.Do()

					str, _ := res.Body.ToString()
					m.Expect(str).Should(m.Equal("foo bar"))
				})

				g.It("Should handle io.ReaderCloser", func() {
					res, _ := Request{Method: "POST", Uri: ts.URL, Body: "foo bar"}.Do()

					body, _ := ioutil.ReadAll(res.Body)
					m.Expect(string(body)).Should(m.Equal("foo bar"))
				})

				g.It("Should handle parsing JSON", func() {
					res, _ := Request{Method: "POST", Uri: ts.URL, Body: `{"foo": "bar"}`}.Do()

					var foobar map[string]string

					res.Body.FromJsonTo(&foobar)

					m.Expect(foobar).Should(m.Equal(map[string]string{"foo": "bar"}))
				})

				g.It("Should return the original request response", func() {
					res, _ := Request{Method: "POST", Uri: ts.URL, Body: `{"foo": "bar"}`}.Do()

					m.Expect(res.Response).ShouldNot(m.BeNil())
				})
			})
			g.Describe("Redirects", func() {
				g.It("Should not follow by default", func() {
					res, _ := Request{
						Uri: ts.URL + "/redirect_test/301",
					}.Do()
					m.Expect(res.StatusCode).Should(m.Equal(301))
				})

				g.It("Should not follow if method is explicitly specified", func() {
					res, err := Request{
						Method: "GET",
						Uri:    ts.URL + "/redirect_test/301",
					}.Do()
					m.Expect(res.StatusCode).Should(m.Equal(301))
					m.Expect(err).ShouldNot(m.HaveOccurred())
				})

				g.It("Should throw an error if MaxRedirect limit is exceeded", func() {
					res, err := Request{
						Method:       "GET",
						MaxRedirects: 1,
						Uri:          ts.URL + "/redirect_test/301",
					}.Do()
					m.Expect(res.StatusCode).Should(m.Equal(302))
					m.Expect(err).Should(m.HaveOccurred())
				})

				g.It("Should copy request headers headers when redirecting if specified", func() {
					req := Request{
						Method:          "GET",
						Uri:             ts.URL + "/redirect_test/301",
						MaxRedirects:    4,
						RedirectHeaders: true,
					}
					req.AddHeader("Testheader", "TestValue")
					res, _ := req.Do()
					m.Expect(res.StatusCode).Should(m.Equal(200))
					m.Expect(requestHeaders.Get("Testheader")).Should(m.Equal("TestValue"))
				})

				g.It("Should follow only specified number of MaxRedirects", func() {
					res, _ := Request{
						Uri:          ts.URL + "/redirect_test/301",
						MaxRedirects: 1,
					}.Do()
					m.Expect(res.StatusCode).Should(m.Equal(302))
					res, _ = Request{
						Uri:          ts.URL + "/redirect_test/301",
						MaxRedirects: 2,
					}.Do()
					m.Expect(res.StatusCode).Should(m.Equal(303))
					res, _ = Request{
						Uri:          ts.URL + "/redirect_test/301",
						MaxRedirects: 3,
					}.Do()
					m.Expect(res.StatusCode).Should(m.Equal(307))
					res, _ = Request{
						Uri:          ts.URL + "/redirect_test/301",
						MaxRedirects: 4,
					}.Do()
					m.Expect(res.StatusCode).Should(m.Equal(200))
				})

				g.It("Should return final URL of the response when redirecting", func() {
					res, _ := Request{
						Uri:          ts.URL + "/redirect_test/destination",
						MaxRedirects: 2,
					}.Do()
					m.Expect(res.Uri).Should(m.Equal(ts.URL + "/destination"))
				})
			})
		})

		g.Describe("Timeouts", func() {

			g.Describe("Connection timeouts", func() {
				g.It("Should connect timeout after a default of 1000 ms", func() {
					start := time.Now()
					res, err := Request{Uri: "http://10.255.255.1"}.Do()
					elapsed := time.Since(start)

					m.Expect(elapsed).Should(m.BeNumerically("<", 1100*time.Millisecond))
					m.Expect(elapsed).Should(m.BeNumerically(">=", 1000*time.Millisecond))
					m.Expect(res).Should(m.BeNil())
					m.Expect(err.(*Error).Timeout()).Should(m.BeTrue())
				})
				g.It("Should connect timeout after a custom amount of time", func() {
					SetConnectTimeout(100 * time.Millisecond)
					start := time.Now()
					res, err := Request{Uri: "http://10.255.255.1"}.Do()
					elapsed := time.Since(start)

					m.Expect(elapsed).Should(m.BeNumerically("<", 150*time.Millisecond))
					m.Expect(elapsed).Should(m.BeNumerically(">=", 100*time.Millisecond))
					m.Expect(res).Should(m.BeNil())
					m.Expect(err.(*Error).Timeout()).Should(m.BeTrue())
				})
				g.It("Should connect timeout after a custom amount of time even with method set", func() {
					SetConnectTimeout(100 * time.Millisecond)
					start := time.Now()
					request := Request{
						Uri:    "http://10.255.255.1",
						Method: "GET",
					}
					res, err := request.Do()
					elapsed := time.Since(start)

					m.Expect(elapsed).Should(m.BeNumerically("<", 150*time.Millisecond))
					m.Expect(elapsed).Should(m.BeNumerically(">=", 100*time.Millisecond))
					m.Expect(res).Should(m.BeNil())
					m.Expect(err.(*Error).Timeout()).Should(m.BeTrue())
				})
			})

			g.Describe("Request timeout", func() {
				var ts *httptest.Server
				stop := make(chan bool)

				g.Before(func() {
					ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						<-stop
						// just wait for someone to tell you when to end the request. this is used to simulate a slow server
					}))
				})
				g.After(func() {
					stop <- true
					ts.Close()
				})
				g.It("Should request timeout after a custom amount of time", func() {
					SetConnectTimeout(1000 * time.Millisecond)

					start := time.Now()
					res, err := Request{Uri: ts.URL, Timeout: 500 * time.Millisecond}.Do()
					elapsed := time.Since(start)

					m.Expect(elapsed).Should(m.BeNumerically("<", 550*time.Millisecond))
					m.Expect(elapsed).Should(m.BeNumerically(">=", 500*time.Millisecond))
					m.Expect(res).Should(m.BeNil())
					m.Expect(err.(*Error).Timeout()).Should(m.BeTrue())
				})
				g.It("Should request timeout after a custom amount of time even with proxy", func() {
					proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						time.Sleep(2000 * time.Millisecond)
						w.WriteHeader(200)
					}))
					SetConnectTimeout(1000 * time.Millisecond)
					start := time.Now()
					request := Request{
						Uri:     ts.URL,
						Proxy:   proxy.URL,
						Timeout: 500 * time.Millisecond,
					}
					res, err := request.Do()
					elapsed := time.Since(start)

					m.Expect(elapsed).Should(m.BeNumerically("<", 550*time.Millisecond))
					m.Expect(elapsed).Should(m.BeNumerically(">=", 500*time.Millisecond))
					m.Expect(res).Should(m.BeNil())
					m.Expect(err.(*Error).Timeout()).Should(m.BeTrue())
				})
			})
		})

		g.Describe("Misc", func() {
			g.It("Should set default golang user agent when not explicitly passed", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					m.Expect(r.Header.Get("User-Agent")).ShouldNot(m.BeZero())
					m.Expect(r.Host).Should(m.Equal("foobar.com"))

					w.WriteHeader(200)
				}))
				defer ts.Close()

				req := Request{Uri: ts.URL, Host: "foobar.com"}
				res, err := req.Do()
				m.Expect(err).ShouldNot(m.HaveOccurred())

				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.It("Should offer to set request headers", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					m.Expect(r.Header.Get("User-Agent")).Should(m.Equal("foobaragent"))
					m.Expect(r.Host).Should(m.Equal("foobar.com"))
					m.Expect(r.Header.Get("Accept")).Should(m.Equal("application/json"))
					m.Expect(r.Header.Get("Content-Type")).Should(m.Equal("application/json"))
					m.Expect(r.Header.Get("X-Custom")).Should(m.Equal("foobar"))
					m.Expect(r.Header.Get("X-Custom2")).Should(m.Equal("barfoo"))

					w.WriteHeader(200)
				}))
				defer ts.Close()

				req := Request{Uri: ts.URL, Accept: "application/json", ContentType: "application/json", UserAgent: "foobaragent", Host: "foobar.com"}
				req.AddHeader("X-Custom", "foobar")
				res, _ := req.WithHeader("X-Custom2", "barfoo").Do()

				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.It("Should call hook before request", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					m.Expect(r.Header.Get("X-Custom")).Should(m.Equal("foobar"))

					w.WriteHeader(200)
				}))
				defer ts.Close()

				hook := func(goreq *Request, httpreq *http.Request) {
					httpreq.Header.Add("X-Custom", "foobar")
				}
				req := Request{Uri: ts.URL, OnBeforeRequest: hook}
				res, _ := req.Do()

				m.Expect(res.StatusCode).Should(m.Equal(200))
			})

			g.It("Should not create a body by defualt", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					b, _ := ioutil.ReadAll(r.Body)
					m.Expect(b).Should(m.HaveLen(0))
					w.WriteHeader(200)
				}))
				defer ts.Close()

				req := Request{Uri: ts.URL, Host: "foobar.com"}
				req.Do()
			})
			g.It("Should change transport TLS config if Request.Insecure is set", func() {
				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				}))
				defer ts.Close()

				req := Request{
					Insecure: true,
					Uri:      ts.URL,
					Host:     "foobar.com",
				}
				res, _ := req.Do()

				m.Expect(DefaultClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify).Should(m.Equal(true))
				m.Expect(res.StatusCode).Should(m.Equal(200))
			})
			g.It("Should work if a different transport is specified", func() {
				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				}))
				defer ts.Close()
				var currentTransport = DefaultTransport
				DefaultTransport = &http.Transport{Dial: DefaultDialer.Dial}

				req := Request{
					Insecure: true,
					Uri:      ts.URL,
					Host:     "foobar.com",
				}
				res, _ := req.Do()

				m.Expect(DefaultClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify).Should(m.Equal(true))
				m.Expect(res.StatusCode).Should(m.Equal(200))

				DefaultTransport = currentTransport

			})
			g.It("GetRequest should return the underlying httpRequest ", func() {
				req := Request{
					Host: "foobar.com",
				}

				request, _ := req.NewRequest()
				m.Expect(request).ShouldNot(m.BeNil())
				m.Expect(request.Host).Should(m.Equal(req.Host))
			})

			g.It("Response should allow to cancel in-flight request", func() {
				unblockc := make(chan bool)
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, "Hello")
					w.(http.Flusher).Flush()
					<-unblockc
				}))
				defer ts.Close()
				defer close(unblockc)

				req := Request{
					Insecure: true,
					Uri:      ts.URL,
					Host:     "foobar.com",
				}
				res, _ := req.Do()
				res.CancelRequest()
				_, err := ioutil.ReadAll(res.Body)
				g.Assert(err != nil).IsTrue()
			})
		})

		g.Describe("Errors", func() {
			var ts *httptest.Server

			g.Before(func() {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == "POST" && r.URL.Path == "/" {
						w.Header().Add("Location", ts.URL+"/123")
						w.WriteHeader(201)
						io.Copy(w, r.Body)
					}
				}))
			})

			g.After(func() {
				ts.Close()
			})
			g.It("Should throw an error when FromJsonTo fails", func() {
				res, _ := Request{Method: "POST", Uri: ts.URL, Body: `{"foo" "bar"}`}.Do()
				var foobar map[string]string

				err := res.Body.FromJsonTo(&foobar)
				m.Expect(err).Should(m.HaveOccurred())
			})
			g.It("Should handle Url parsing errors", func() {
				_, err := Request{Uri: ":"}.Do()

				m.Expect(err).ShouldNot(m.BeNil())
			})
			g.It("Should handle DNS errors", func() {
				_, err := Request{Uri: "http://.localhost"}.Do()
				m.Expect(err).ShouldNot(m.BeNil())
			})
		})

		g.Describe("Proxy", func() {
			var ts *httptest.Server
			var lastReq *http.Request
			g.Before(func() {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == "GET" && r.URL.Path == "/" {
						lastReq = r
						w.Header().Add("x-forwarded-for", "test")
						w.Header().Add("Set-Cookie", "foo=bar")
						w.WriteHeader(200)
						w.Write([]byte(""))
					} else if r.Method == "GET" && r.URL.Path == "/redirect_test/301" {
						http.Redirect(w, r, "/", 301)
					} else if r.Method == "CONNECT" {
						lastReq = r
					}
				}))

			})

			g.BeforeEach(func() {
				lastReq = nil
			})

			g.After(func() {
				ts.Close()
			})

			g.It("Should use Proxy", func() {
				proxiedHost := "www.google.com"
				res, err := Request{Uri: "http://" + proxiedHost, Proxy: ts.URL}.Do()
				m.Expect(err).Should(m.BeNil())
				m.Expect(res.Header.Get("x-forwarded-for")).Should(m.Equal("test"))
				m.Expect(lastReq).ShouldNot(m.BeNil())
				m.Expect(lastReq.Host).Should(m.Equal(proxiedHost))
			})

			g.It("Should not redirect if MaxRedirects is not set", func() {
				res, err := Request{Uri: ts.URL + "/redirect_test/301", Proxy: ts.URL}.Do()
				m.Expect(err).ShouldNot(m.HaveOccurred())
				m.Expect(res.StatusCode).Should(m.Equal(301))
			})

			g.It("Should use Proxy authentication", func() {
				proxiedHost := "www.google.com"
				uri := strings.Replace(ts.URL, "http://", "http://user:pass@", -1)
				res, err := Request{Uri: "http://" + proxiedHost, Proxy: uri}.Do()
				m.Expect(err).Should(m.BeNil())
				m.Expect(res.Header.Get("x-forwarded-for")).Should(m.Equal("test"))
				m.Expect(lastReq).ShouldNot(m.BeNil())
				m.Expect(lastReq.Header.Get("Proxy-Authorization")).Should(m.Equal("Basic dXNlcjpwYXNz"))
			})

			g.It("Should propagate cookies", func() {
				proxiedHost, _ := url.Parse("http://www.google.com")
				jar, _ := cookiejar.New(nil)
				res, err := Request{Uri: proxiedHost.String(), Proxy: ts.URL, CookieJar: jar}.Do()
				m.Expect(err).Should(m.BeNil())
				m.Expect(res.Header.Get("x-forwarded-for")).Should(m.Equal("test"))

				m.Expect(jar.Cookies(proxiedHost)).Should(m.HaveLen(1))
				m.Expect(jar.Cookies(proxiedHost)[0].Name).Should(m.Equal("foo"))
				m.Expect(jar.Cookies(proxiedHost)[0].Value).Should(m.Equal("bar"))
			})

			g.It("Should use ProxyConnectHeader authentication", func() {
				_, err := Request{Uri: "https://10.255.255.1",
					Proxy:    ts.URL,
					Insecure: true,
				}.WithProxyConnectHeader("X-TEST-HEADER", "TEST").Do()

				m.Expect(err).ShouldNot(m.BeNil())
				m.Expect(lastReq.Header.Get("X-TEST-HEADER")).Should(m.Equal("TEST"))
			})

		})

		g.Describe("BasicAuth", func() {
			var ts *httptest.Server

			g.Before(func() {
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/basic_auth" {
						auth_array := r.Header["Authorization"]
						if len(auth_array) > 0 {
							auth := strings.TrimSpace(auth_array[0])
							w.WriteHeader(200)
							fmt.Fprint(w, auth)
						} else {
							w.WriteHeader(401)
							fmt.Fprint(w, "private")
						}
					}
				}))

			})

			g.After(func() {
				ts.Close()
			})

			g.It("Should support basic http authorization", func() {
				res, err := Request{
					Uri:               ts.URL + "/basic_auth",
					BasicAuthUsername: "username",
					BasicAuthPassword: "password",
				}.Do()
				m.Expect(err).Should(m.BeNil())
				str, _ := res.Body.ToString()
				m.Expect(res.StatusCode).Should(m.Equal(200))
				expectedStr := "Basic " + base64.StdEncoding.EncodeToString([]byte("username:password"))
				m.Expect(str).Should(m.Equal(expectedStr))
			})

			g.It("Should fail when basic http authorization is required and not provided", func() {
				res, err := Request{
					Uri: ts.URL + "/basic_auth",
				}.Do()
				m.Expect(err).Should(m.BeNil())
				str, _ := res.Body.ToString()
				m.Expect(res.StatusCode).Should(m.Equal(401))
				m.Expect(str).Should(m.Equal("private"))
			})
		})
	})
}

func Test_paramParse(t *testing.T) {
	type Form struct {
		A string
		B string
		c string
	}

	type AnnotedForm struct {
		Foo  string `url:"foo_bar"`
		Baz  string `url:"bad,omitempty"`
		Norf string `url:"norf,omitempty"`
		Qux  string `url:"-"`
	}

	type EmbedForm struct {
		AnnotedForm `url:",squash"`
		Form        `url:",squash"`
		Corge       string `url:"corge"`
	}

	g := gob.Goblin(t)
	m.RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	var form = Form{}
	var aform = AnnotedForm{}
	var eform = EmbedForm{}
	var values = url.Values{}
	const result = "a=1&b=2"
	g.Describe("QueryString ParamParse", func() {
		g.Before(func() {
			form.A = "1"
			form.B = "2"
			form.c = "3"
			aform.Foo = "xyz"
			aform.Norf = "abc"
			aform.Qux = "def"
			eform.Form = form
			eform.AnnotedForm = aform
			eform.Corge = "xxx"
			values.Add("a", "1")
			values.Add("b", "2")
		})
		g.It("Should accept struct and ignores unexported field", func() {
			str, err := paramParse(form)
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal(result))
		})
		g.It("Should accept struct and use the field annotations", func() {
			str, err := paramParse(aform)
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal("foo_bar=xyz&norf=abc"))
		})
		g.It("Should accept pointer of struct", func() {
			str, err := paramParse(&form)
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal(result))
		})
		g.It("Should accept recursive pointer of struct", func() {
			f := &form
			ff := &f
			str, err := paramParse(ff)
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal(result))
		})
		g.It("Should accept embedded struct", func() {
			str, err := paramParse(eform)
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal("a=1&b=2&corge=xxx&foo_bar=xyz&norf=abc"))
		})
		g.It("Should accept interface{} which forcely converted by struct", func() {
			str, err := paramParse(interface{}(&form))
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal(result))
		})

		g.It("Should accept url.Values", func() {
			str, err := paramParse(values)
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal(result))
		})
		g.It("Should accept &url.Values", func() {
			str, err := paramParse(&values)
			m.Expect(err).Should(m.BeNil())
			m.Expect(str).Should(m.Equal(result))
		})
	})

}
