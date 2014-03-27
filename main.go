package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"

	"./transform"
)

import _ "net/http/pprof"

var g_transform transform.Transform
var SiteConfig = map[string]interface{}{}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	home, err := template.ParseFiles("./template/home.html")
	if err != nil {
		panic(err)
	}
	home.Execute(w, SiteConfig)
}

func HookHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "javascript/hook.js" {
		http.NotFound(w, r)
		return
	}
	w.Header().Add("Content-Type", "text/javascript")

	hookjs, err := template.ParseFiles("./javascript/hook.js")
	if err != nil {
		panic(err)
	}
	hookjs.Execute(w, SiteConfig)
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	source_url := r.FormValue("url")
	if !strings.HasPrefix(source_url, "http://") &&
		!strings.HasPrefix(source_url, "https://") {
		source_url = "http://" + source_url
	}

	dest_url := g_transform.ProcessLink("/db.a/", source_url)

	http.Redirect(w, r, dest_url, 301)
}

func TravelHandler(w http.ResponseWriter, r *http.Request) {
	base_url := r.URL.Path[len("/access/"):]
	// process url
	rindex := strings.LastIndex(base_url, "/")
	if rindex != -1 {
		base_url = base_url[:rindex+1]
	}

	err := transform.ModifyRequest(r, &g_transform)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 404)
		return
	}
	fmt.Println("REQ_URL:", r.URL)

	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	resp, err := tr.RoundTrip(r)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 404)
		return
	}

	// fmt.Println("RESP_CODE: ", resp.StatusCode)

	body := transform.ModifyRespBody(resp, base_url, &g_transform)
	transform.ModifyResponse(resp, base_url, &g_transform)

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	buffer, err := ioutil.ReadFile("./config/site.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(buffer, &SiteConfig)
	SiteConfig["RootURI"] = SiteConfig["Protocol"].(string) + "://" +
		SiteConfig["AccessAddress"].(string)
	SiteConfig["FullURI"] = SiteConfig["RootURI"].(string) +
		SiteConfig["AccessPath"].(string)

	err = g_transform.Init(SiteConfig["FullURI"].(string))
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/access/", TravelHandler)
	http.HandleFunc("/javascript/", HookHandler)
	http.HandleFunc("/proxy.php", ProxyHandler)

	http.ListenAndServe(SiteConfig["ListenAddress"].(string), nil)
}
