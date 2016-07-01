package transform

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func ModifyRequest(r *http.Request, t *Transform) error {
	req_uri := r.URL.RequestURI()[len("/access/"):]
	req_uri = t.DecodeURI(req_uri)

	u, err := url.Parse(req_uri)
	if err != nil {
		return err
	}
	r.URL = u
	r.Host = u.Host

	r.Header.Del("Host")
	r.Header.Del("Accept-Encoding")

	if r.Header.Get("Referer") != "" {
		refer := t.DecodeAddress(r.Header.Get("Referer"))
		r.Header.Set("Referer", refer)
	}
	return nil
}

func ModifyResponse(resp *http.Response, base_url string, t *Transform) error {
	if resp.Header.Get("Location") != "" {
		// fmt.Println("Location:", resp.Header.Get("Location"))

		// loc := t.EncodeURI(resp.Header.Get("Location"))
		loc := t.ProcessLink(base_url, resp.Header.Get("Location"))
		resp.Header.Set("Location", loc)

		// fmt.Println("Location final:", loc)
	}

	resp.Header.Del("Content-Length")
	resp.Header.Del("Content-Encoding")
	resp.Header.Del("Access-Control-Allow-Origin")
	resp.Header.Del("X-WebKit-CSP")
	resp.Header.Del("Content-Security-Policy")

	// fmt.Println("SET-COOKIE", resp.Cookies())

	cookies := resp.Cookies()
	resp.Header.Del("Set-Cookie")
	for _, v := range cookies {
		v.Path = "/access/"
		if v.Domain != "" {
			if v.Domain[0] == '.' {
				v.Domain = v.Domain[1:]
			} else if strings.HasPrefix(v.Domain, "*.") {
				v.Domain = v.Domain[2:]
			}
			v.Domain = t.ReverseHost(v.Domain, ".", "/")
			v.Path = "/access/" + v.Domain + "/"
		}

		v.Domain = ""
		v.Secure = false
		resp.Header.Add("Set-Cookie", v.String())
	}
	return nil
}

func ModifyRespBody(resp *http.Response, base_url string, t *Transform) error {
	ctype := ""
	var body []byte

	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		ctype = "html"
	} else if strings.Contains(resp.Header.Get("Content-Type"), "xml") {
		ctype = "xml"
	} else if strings.Contains(resp.Header.Get("Content-Type"), "text/css") {
		ctype = "css"
	} else if strings.Contains(resp.Header.Get("Content-Type"), "javascript") {
		ctype = "javascript"
	} else {
		return nil
	}

	defer resp.Body.Close()

	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Println("GZIP: ", err)
			body, err = ioutil.ReadAll(resp.Body)
		} else {
			body, err = ioutil.ReadAll(gr)
		}
	} else {
		var err error
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("READ: ", err)
		}
	}

	switch ctype {
	case "html", "xml":
		html := string(body)
		// fmt.Println("BEFORE_BODY", html)
		html, _ = t.ProcessHTML(base_url, html)
		// fmt.Println("AFTER_BODY", html)
		body = []byte(html)
	case "css":
		body, _ = t.ProcessCSS(base_url, body)
	case "javascript":
		body, _ = t.ProcessJS(base_url, body)
	}

	resp.Body = ioutil.NopCloser(bytes.NewReader(body))

	return nil
}
