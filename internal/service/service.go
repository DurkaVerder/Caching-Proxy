package service

import (
	"Caching-Proxy/internal/cache"
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

type Service interface {
	Run()
}

type ServiceManager struct {
	cache *cache.Cache
	Port  string
	URL   string
}

func NewServiceManager(args []string, cache *cache.Cache) *ServiceManager {
	port, URL, err := parserArgs(args)
	if err != nil {
		log.Fatalf("Error parsing args: %v", err)
	}
	return &ServiceManager{
		Port:  ":" + port,
		URL:   URL,
		cache: cache,
	}
}

func parserArgs(args []string) (port string, URL string, err error) {
	if len(args) < 5 {
		return "", "", errors.New("not enough arguments")
	}

	if args[1] != "--port" {
		return "", "", errors.New("invalid argument " + args[1])
	}
	if args[3] != "--origin" {
		return "", "", errors.New("invalid argument " + args[3])
	}

	port = args[2]
	for _, c := range port {
		if c < '0' || c > '9' {
			return "", "", errors.New("invalid port number")
		}
	}

	URL = args[4]

	return port, URL, nil
}

func (s *ServiceManager) Run() {
	log.Printf("Starting server on port %s with origin %s", s.Port, s.URL)

	http.HandleFunc("/", s.cacheHandler)
	http.ListenAndServe(s.Port, nil)
}

func (s *ServiceManager) cacheHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	targetURL := s.URL + r.URL.Path
	query := r.URL.RawQuery
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	key := s.createKey(method, targetURL, query, string(body))
	if method == http.MethodGet {
		if value, found := s.cache.Get(key); found {
			w.Header().Add("X-CACHE", "HIT")
			w.Write(value.([]byte))
			return
		}
	}

	resp, err := s.requestToServer(method, targetURL, query, body)
	if err != nil {
		http.Error(w, "Error requesting to server", http.StatusInternalServerError)
		return
	}

	respBody, err := s.getResponseBody(resp)
	if err != nil {
		http.Error(w, "Error getting response body", http.StatusInternalServerError)
		return
	}

	if method == http.MethodGet {
		s.cache.Set(key, respBody, 0)
	}

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.Header().Add("X-CACHE", "MISS")

	w.Write(respBody)

}

func (s *ServiceManager) createKey(method, targetURL, query, body string) string {
	b := strings.Builder{}
	b.WriteString(method)
	b.WriteString(targetURL)
	b.WriteString(query)
	b.WriteString(body)
	return b.String()
}

func (s *ServiceManager) requestToServer(method, targetURL, query string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *ServiceManager) getResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
