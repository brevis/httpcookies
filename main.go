package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type CookieInfo struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  string `json:"expires"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type HeaderInfo struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type Response struct {
	RequestedURL string       `json:"requested_url"`
	Method       string       `json:"method"`
	Status       string       `json:"status"`
	Cookies      []CookieInfo `json:"cookies,omitempty"`
	Headers      []HeaderInfo `json:"headers,omitempty"`
	Error        string       `json:"error,omitempty"`
}

type spaHandler struct {
	staticPath string
	indexPath  string
}

func isValidURL(targetURL string) (bool, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false, fmt.Errorf("invalid url")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false, fmt.Errorf("unsupported scheme")
	}

	if strings.HasPrefix(parsedURL.Host, "localhost") || strings.Contains(parsedURL.Host, "127.0.0.1") || strings.Contains(parsedURL.Host, "::1") {
		return false, fmt.Errorf("unsupported host")
	}

	return true, nil
}

func isValidMethod(method string) bool {
	methods := []string{"GET", "POST", "PUT", "PATCH", "HEAD", "DELETE", "OPTIONS", "TRACE", "CONNECT"}
	return slices.Contains(methods, strings.ToUpper(method))
}

func fetchCustomRequest(targetURL, method string, headers map[string]string, body string) (*Response, error) {
	valid, err := isValidURL(targetURL)
	if !valid {
		return nil, err
	}

	parsedURL, _ := url.Parse(targetURL)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	var bodyReader io.Reader = nil
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error create http request: %v", err)
	}

	if !isValidMethod(method) {
		return nil, fmt.Errorf("unsupported http method")
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error make http request: %v", err)
	}
	defer resp.Body.Close()

	var cookieData []CookieInfo
	for _, c := range jar.Cookies(parsedURL) {
		cookieData = append(cookieData, CookieInfo{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  c.Expires.String(),
			HttpOnly: c.HttpOnly,
			Secure:   c.Secure,
		})
	}

	var headerData []HeaderInfo
	for name, values := range resp.Header {
		headerData = append(headerData, HeaderInfo{
			Name:   name,
			Values: values,
		})
	}

	return &Response{
		RequestedURL: targetURL,
		Method:       method,
		Status:       resp.Status,
		Cookies:      cookieData,
		Headers:      headerData,
	}, nil
}

func ParseHeaders(headers string) map[string]string {
	headerMap := make(map[string]string)
	lines := strings.Split(headers, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headerMap[key] = value
	}

	return headerMap
}

func getCookiesHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	method := r.FormValue("method")
	body := r.FormValue("body")
	headers := ParseHeaders(r.FormValue("headers"))

	result, err := fetchCustomRequest(url, method, headers, body)
	if err != nil {
		result = &Response{
			RequestedURL: url,
			Method:       method,
			Error:        err.Error(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.staticPath, r.URL.Path)
	fi, err := os.Stat(path)
	if os.IsNotExist(err) || fi.IsDir() {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	r := mux.NewRouter()

	r.PathPrefix("/").Handler(spaHandler{
		staticPath: "static",
		indexPath:  "index.html",
	}).Methods("GET")

	r.HandleFunc("/get-cookies", getCookiesHandler).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:" + port,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
