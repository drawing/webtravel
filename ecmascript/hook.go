package ecmascript

import (
	"bytes"
)

type calleeConfig struct {
	// 0:match 1:suffix
	matchType    int
	matchText    []byte
	argumentNum  int
	replaceNum   int
	replaceFunc  []byte
	argumentType int
}

var calleConfigList = []*calleeConfig{
	// facebook.org
	&calleeConfig{0, []byte("window.location.replace"), 1, 0, []byte("skip_access_convert_url_0001"), 0},
	&calleeConfig{0, []byte("location.replace"), 1, 0, []byte("skip_access_convert_url_0001"), 0},
	// play.golang.org
	&calleeConfig{1, []byte(".pushState"), 3, 2, []byte("skip_access_convert_path_0001"), 0},
	&calleeConfig{1, []byte(".setAttribute"), 2, 1, []byte("skip_access_setAttribute_0001"), 1},
	&calleeConfig{1, []byte(".test"), 1, 0, []byte("skip_access_regex_test_0001"), 0},
}

func calleeHook(text []byte, num int) ([]byte, int, int) {
	match := false
	for _, v := range calleConfigList {
		if num != v.argumentNum {
			continue
		}
		if v.matchType == 0 {
			if bytes.Equal(text, v.matchText) {
				match = true
			}
		} else if v.matchType == 1 {
			if bytes.HasSuffix(text, v.matchText) {
				match = true
			}
		}
		if !match {
			continue
		}
		return v.replaceFunc, v.replaceNum, v.argumentType
	}

	return nil, 0, 0
}

var memberConfigList = map[string][]byte{
	"window.location.host":     []byte("skip_access_convert_host_0001"),
	"window.location.protocol": []byte("skip_access_convert_protocol_0001"),
}

type assignmentConfig struct {
	// 0:match 1:suffix
	matchType   int
	matchText   []byte
	replaceFunc []byte
}

var assignmentConfigList = []*assignmentConfig{
	&assignmentConfig{1, []byte(".href"), []byte("skip_access_convert_url_0001")},
	&assignmentConfig{1, []byte(".src"), []byte("skip_access_convert_url_0001")},
	&assignmentConfig{1, []byte(".domain"), []byte("skip_access_convert_domain_0001")},
	&assignmentConfig{1, []byte(".innerHTML"), []byte("skip_access_convert_html_0001")},
	&assignmentConfig{1, []byte(".cookie"), []byte("skip_access_convert_cookie_0001")},
	&assignmentConfig{1, []byte(".outerHTML"), []byte("skip_access_convert_html_0001")},
}

func assignmentHook(left []byte) []byte {
	var hook []byte = nil
	for _, v := range assignmentConfigList {
		if v.matchType == 0 {
			if bytes.Equal(left, v.matchText) {
				hook = v.replaceFunc
			}
		} else if v.matchType == 1 {
			if bytes.HasSuffix(left, v.matchText) {
				hook = v.replaceFunc
			}
		}
		if hook != nil {
			return hook
		}
	}

	return nil
}
