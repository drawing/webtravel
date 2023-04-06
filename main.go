package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
	"flag"

	"github.com/drawing/webtravel/transform"
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

func RobotsHandler(w http.ResponseWriter, r *http.Request) {
	robots, err := template.ParseFiles("./template/robots.txt")
	if err != nil {
		panic(err)
	}
	robots.Execute(w, SiteConfig)
}

func HookHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/javascript/hook.js" {
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
		log.Println(err)
		http.Redirect(w, r, "/", 404)
		return
	}
	log.Println("REQ_URL:", r.URL)

	tr := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	resp, err := tr.RoundTrip(r)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", 404)
		return
	}

	// log.Println("RESP_CODE: ", resp.StatusCode)

	err = transform.ModifyRespBody(resp, base_url, &g_transform)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", 500)
		return
	}

	defer resp.Body.Close()

	transform.ModifyResponse(resp, base_url, &g_transform)

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	ConfigFile := flag.String("config", "./config/site.json", "Config File Path")

	flag.Parse()

	log.Println("read config file:", *ConfigFile)

	buffer, err := ioutil.ReadFile(*ConfigFile)
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
		log.Println(err)
		return
	}

	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/robots.txt", RobotsHandler)
	http.HandleFunc("/access/", TravelHandler)
	http.HandleFunc("/javascript/", HookHandler)
	http.HandleFunc("/proxy.php", ProxyHandler)

	log.Println("Start Running @", SiteConfig["ListenAddress"], "...")
	http.ListenAndServe(SiteConfig["ListenAddress"].(string), nil)
}
