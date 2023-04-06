package transform

import (
	"bytes"
	"errors"
	"strings"
)

import (
	"golang.org/x/net/html"
)
import (
	"github.com/drawing/webtravel/ecmascript"
)

var c_link_attrs = map[string]bool{
	"action": true, "archive": true, "background": true, "cite": true, "classid": true,
	"codebase": true, "data": true, "href": true, "longdesc": true, "profile": true, "src": true,
	"usemap": true,
	"sprite": true,
	// Not standard:
	"dynsrc": true, "lowsrc": true,
}

type Transform struct {
	site_url string
	css      CSSTransform
	js       ecmascript.JSTransform
}

func (t *Transform) Init(site_url string) error {
	t.site_url = site_url
	return t.css.Init(t)
}

// base_url must is "a/" prefix or "/" suffix
func (t *Transform) ProcessLink(base_url string, url string) string {
	// log.Println("process", base_url, url)
	var result []string
	length := len(url)
	if length == 0 {
		return ""
	}

	if strings.HasPrefix(url, t.site_url) {
		return url
	}

	var rep string = ""
	if url[0] == '\'' || url[length-1] == '"' {
		if length <= 2 {
			return url
		}
		rep = url[0:1]
		result = append(result, rep)
		url = url[1 : length-1]
	}

	pos := strings.Index(base_url, "/db.")
	if pos == -1 {
		return url
	}

	for strings.HasPrefix(url, "../") {
		url = url[3:]
		spos := strings.LastIndex(base_url[:len(base_url)-1], "/")
		if spos != -1 && spos != pos {
			base_url = base_url[:spos+1]
		}
		if len(url) == 0 {
			return t.site_url + base_url
		}
	}

	if strings.HasPrefix(url, "#") {
		result = append(result, url)
	} else if strings.HasPrefix(url, "about:blank") {
		result = append(result, url)
	} else if strings.HasPrefix(url, "data:") {
		result = append(result, url)
	} else if strings.HasPrefix(url, "file:") {
		result = append(result, url)
	} else if strings.HasPrefix(url, "res:") {
		result = append(result, url)
	} else if strings.HasPrefix(url, "C:") {
		result = append(result, url)
	} else if strings.HasPrefix(url, "javascript:") {
		result = append(result, url)
	} else if strings.HasPrefix(url, t.site_url) {
		result = append(result, url)
	} else if strings.HasPrefix(url, "//") {
		host, uri := t.SplitAddress(url[2:])
		result = append(result, t.site_url)
		result = append(result, t.ReverseHost(host, ".", "/"))
		if strings.Contains(base_url, "db.b") {
			result = append(result, "/db.b")
		} else {
			result = append(result, "/db.a")
		}
		result = append(result, uri)
	} else if url[0] == '/' {
		var i = strings.Index(base_url, "db.")
		if i == -1 {
			i = len(base_url)
		} else {
			i += 4
		}

		result = append(result, t.site_url)
		result = append(result, base_url[:i])
		result = append(result, url)
	} else if strings.HasPrefix(url, "http://") {
		host, uri := t.SplitAddress(url[7:])
		result = append(result, t.site_url)
		result = append(result, t.ReverseHost(host, ".", "/"))
		result = append(result, "/db.a")
		result = append(result, uri)

	} else if strings.HasPrefix(url, "https://") {
		host, uri := t.SplitAddress(url[8:])
		result = append(result, t.site_url)
		result = append(result, t.ReverseHost(host, ".", "/"))
		result = append(result, "/db.b")
		result = append(result, uri)
	} else {
		result = append(result, t.site_url)
		result = append(result, base_url)
		result = append(result, url)
	}

	// result = t.site_url + base_url + "converted"

	if rep != "" {
		// result = strings.Join([]string{rep, result, rep}, "")
		result = append(result, rep)
	}

	// fmt.Println("Process", base_url, url, strings.Join(result, ""))
	return strings.Join(result, "")
}

func (t *Transform) ProcessHTML(base_url string, text string) (string, error) {
	isHtml := false
	for i := 0; i < len(text); i += 1 {
		if text[i] == '\t' || text[i] == '\n' ||
			text[i] == ' ' || text[i] == '\r' {
			continue
		}
		if text[i] == '<' {
			isHtml = true
		}
		break
	}
	if !isHtml {
		return text, errors.New("html format error")
	}

	doc, err := html.Parse(strings.NewReader(text))
	if err != nil {
		return text, err
	}
	t.walk(base_url, doc)

	// modify header
	outter := doc.FirstChild
	for outter != nil && outter.Type != html.ElementNode {
		outter = outter.NextSibling
	}
	if outter != nil && outter.Data == "html" {
		parent := outter.FirstChild
		for parent != nil && parent.Type != html.ElementNode {
			parent = parent.NextSibling
		}
		// fmt.Println("child", parent)
		if parent != nil && parent.Data == "head" {
			new_node := &html.Node{
				Type: html.ElementNode,
				Data: "script",
				Attr: []html.Attribute{
					html.Attribute{"", "type", "text/javascript"},
					html.Attribute{"", "src", "/javascript/hook.js"},
				},
			}
			parent.InsertBefore(new_node, parent.FirstChild)
		}
	}

	b := new(bytes.Buffer)

	if err = html.Render(b, doc); err != nil {
		return text, err
	}

	return b.String(), nil
}

func (t *Transform) ProcessCSS(base_url string, text []byte) ([]byte, error) {
	css, err := t.css.Process(base_url, text)
	return css, err
}
func (t *Transform) ProcessJS(base_url string, text []byte) ([]byte, error) {
	comment := false
	text = bytes.TrimSpace(text)
	if len(text) > 7 &&
		bytes.HasPrefix(text, []byte("<!--")) &&
		bytes.HasSuffix(text, []byte("-->")) {
		comment = true
		text = text[4 : len(text)-3]
	}
	p, err := t.js.Process(text)
	if err != nil {
		return text, err
	}
	if comment {
		// p = append([]byte("<!--"), p)
		// p = append(p, []byte("-->"))
	}
	return p, nil
}

func (t *Transform) writeCSS(base_url string, n *html.Node) error {
	if n == nil || n.Type != html.TextNode {
		return nil
	}
	b, _ := t.ProcessCSS(base_url, []byte(n.Data))
	n.Data = string(b)
	return nil
}
func (t *Transform) writeJS(base_url string, n *html.Node) error {
	if n == nil || n.Type != html.TextNode {
		return nil
	}
	b, _ := t.ProcessJS(base_url, []byte(n.Data))
	n.Data = string(b)
	return nil
}

func (t *Transform) writeObject(base_url string, attributes []html.Attribute) {
}

func (t *Transform) writeMeta(base_url string, attributes []html.Attribute) {
	// fmt.Println("meta")
	var change string = ""
	for k, v := range attributes {
		if v.Key == change {
			pos := strings.Index(strings.ToLower(v.Val), "url=")
			if pos == -1 {
				return
			}
			v.Val = v.Val[:pos+4] + t.ProcessLink(base_url, v.Val[pos+4:])
			attributes[k].Val = v.Val
			// fmt.Println("meta:", v.Val)
		} else if strings.ToLower(v.Key) == "http-equiv" &&
			strings.ToLower(v.Val) == "refresh" {
			change = "content"
			// fmt.Println("meta:", change)
		}
	}
}

func (t *Transform) writeAttribute(base_url string, attributes []html.Attribute) {
	// fmt.Println("Attrs: ", attributes)
	for k, v := range attributes {
		if v.Key == "style" {
			b, _ := t.ProcessCSS(base_url, []byte(v.Val))
			v.Val = string(b)
			continue
		} else if v.Key == "script" {
			b, _ := t.ProcessJS(base_url, []byte(v.Val))
			v.Val = string(b)
			continue
		}

		_, present := c_link_attrs[v.Key]
		// fmt.Println("AttrKey: ", "-"+v.Key+"-", present)
		if present {
			attributes[k].Val = t.ProcessLink(base_url, v.Val)
		}
		// fmt.Println("AttrValue: ", "-"+v.Val+"-", present)
	}
	// fmt.Println("AttrsAfter: ", attributes)
}

// /usr/lib/python2.7/dist-packages/lxml/html/__init__.py:322
func (t *Transform) walk(base_url string, n *html.Node) {
	// fmt.Println(n.Data)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.walk(base_url, c)
	}

	if n.Type != html.ElementNode {
		return
	}

	if n.Data == "meta" {
		t.writeMeta(base_url, n.Attr)
	}

	if n.Data == "object" {
		t.writeObject(base_url, n.Attr)
	} else {
		t.writeAttribute(base_url, n.Attr)
	}

	switch n.Data {
	case "param":
		// rewrite param
	case "style":
		t.writeCSS(base_url, n.FirstChild)
	case "script":
		t.writeJS(base_url, n.FirstChild)
	}
}

func (t *Transform) DecodeAddress(url string) string {
	if !strings.HasPrefix(url, t.site_url) {
		return url
	}
	url = url[len(t.site_url):]

	return t.DecodeURI(url)
}

func (t *Transform) DecodeURI(url string) string {
	pos := strings.Index(url, "db.")
	if pos <= 1 {
		return url
	}
	host := url[:pos-1]
	url = url[pos+3:]

	host = t.ReverseHost(host, "/", ".")

	if strings.HasPrefix(url, "a/") {
		url = "http://" + host + url[1:]
	} else if strings.HasPrefix(url, "b/") {
		url = "https://" + host + url[1:]
	} else {
		url = "http://" + host + "/" + url
	}
	return url
}

func (t *Transform) ReverseHost(host string, begin string, end string) string {
	arr := strings.Split(host, begin)
	i := 0
	j := len(arr) - 1
	for i < j {
		temp := arr[i]
		arr[i] = arr[j]
		arr[j] = temp
		i += 1
		j -= 1
	}
	host = strings.Join(arr, end)
	return host
}
func (t *Transform) SplitAddress(url string) (string, string) {
	pos := strings.Index(url, "/")
	if pos == -1 {
		return url, "/"
	}
	return url[:pos], url[pos:]
}
