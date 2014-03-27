package transform

import "testing"

func TestJavaScript(t *testing.T) {
	TransformInit("https://web2live.com/access")

	html := "function a(){b=12;alert(b)}"
	result := TransformHTML(html, "http://www.baidu.com", "javascript")
	t.Log(result)
	TransformDestroy()
}
