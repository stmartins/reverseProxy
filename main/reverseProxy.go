package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type requestPayloadStruct struct {
	ProxyCondition string `json:"proxy_condition"`
}

func getEnv(key, fallback string) string {

	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getListenAddress() string {
	port := getEnv("PORT", "1338")
	return ":" + port
}

func logSetup() {
	a_condition_url := os.Getenv("A_CONDITION_URL")
	b_condition_url := os.Getenv("B_CONDITION_URL")
	default_condition_url := os.Getenv("DEFAULT_CONDITION_URL")

	log.Println("Server will run on port", getListenAddress())
	log.Printf("Redirecting to A url: %s\n", a_condition_url)
	log.Printf("Redirecting to B url: %s\n", b_condition_url)
	log.Printf("Redirecting to Default url: %s\n", default_condition_url)
}

// Get a json decoder for a given requests body
func requestBodyDocoder(req *http.Request) *json.Decoder {
	// Read body to buffer
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		panic(err)
	}
	// Because go lang is a pain in the ass if you read the body then any susequent calls
	// are unable to read the body again....
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
}

func parseRequestBody(req *http.Request) requestPayloadStruct {
	decoder := requestBodyDocoder(req)
	var requestPayload requestPayloadStruct

	err := decoder.Decode(&requestPayload)
	if err != nil {
		panic(err)
	}
	return requestPayload
}

func logRequestPayload(requestPayload requestPayloadStruct, proxyUrl string) {
	log.Printf("proxy_condition: %s, proxy_url: %s\n", requestPayload.ProxyCondition, proxyUrl)
}

func getProxyURL(proxyConditionRaw string) string {
	proxyCondition := strings.ToUpper(proxyConditionRaw)

	a_condition_url := os.Getenv("A_CONDITION_URL")
	b_condition_url := os.Getenv("B_CONDITION_URL")
	default_condition_url := os.Getenv("DEFAULT_CONDITION_URL")

	if proxyCondition == "A" {
		return a_condition_url
	} else if proxyCondition == "B" {
		return b_condition_url
	}
	return default_condition_url
}

// Serve a reverse proxy for a given url
func serverReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	url, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	log.Println("req.URL.Hos:", req.URL.Host)
	log.Println("req.URL.Scheme:", req.URL.Scheme)
	log.Println("req.Host:", req.Host)

	proxy.ServeHTTP(res, req)
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	requestPayload := parseRequestBody(req)
	url := getProxyURL(requestPayload.ProxyCondition)

	logRequestPayload(requestPayload, url)

	serverReverseProxy(url, res, req)
}

func main() {
	logSetup()

	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
}
