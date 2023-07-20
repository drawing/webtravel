package transform

import "testing"

func TestJavaScript(t *testing.T) {
	var transform Transform
	transform.Init("https://web2live.com/access")
	
	html := []byte("function jump(url){window.location.href = url;}")
	var result, err = transform.ProcessJS("http://www.baidu.com", html)
	if (err != nil) {
		t.Fatal("process js:", err)
	}
	var compared = "function jump(url){window.location.href=skip_access_convert_url_0001(url);}"
	if (string(result) != compared) {
		t.Fatal("process js not equal:", result)
	}
}
