package transform

import (
	"bytes"
	"regexp"
	"strings"
	// "fmt"
)

type CSSTransform struct {
	url_exp *regexp.Regexp
	imp_exp *regexp.Regexp

	link_trans *Transform
}

func (t *CSSTransform) Init(trans *Transform) error {
	var err error
	t.url_exp, err = regexp.Compile("url\\(([\"][^\"]*[\"]|['][^']*[']|[^)]*)\\)")
	if err != nil {
		return err
	}
	t.imp_exp, err = regexp.Compile("@import \"(.*?)\"")
	if err != nil {
		return err
	}

	t.link_trans = trans
	return nil
}

func (t *CSSTransform) Process(base_url string, text []byte) ([]byte, error) {
	res := t.url_exp.ReplaceAllFunc(text, func(link []byte) []byte {
		length := len(link)
		if length == 0 {
			return []byte("")
		}

		// fmt.Println("css", base_url, "-" + string(link) + "-")
		var ll string
		if bytes.HasPrefix(link, []byte("url(")) && link[length-1] == ')' {
			// fmt.Println("css__prefix", base_url, "-" + string(link) + "-")
			ll = t.link_trans.ProcessLink(base_url, string(link[len("url("):length-1]))
			return []byte(strings.Join([]string{"url(", ll, ")"}, ""))
		}

		ll = t.link_trans.ProcessLink(base_url, string(link))
		return []byte(ll)
	})
	return res, nil
}
