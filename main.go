package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
)

const (
	filePath = "url_map.json"
)

/*
Test
curl -X POST http://127.0.0.1:8080/google?dst=https%3a%2f%2fwww%2egoogle%2ecom%2f
curl --dump-header - http://127.0.0.1:8080/google
*/

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetURL(w, r)
		case http.MethodPost:
			PostURL(w, r)
		}
	})
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// GetURL [GET] /
func GetURL(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Path

	url := getRedirectURL(src)
	if url == nil || *url == "" {
		return
	}

	log.Print("redirect:" + *url)

	w.Header().Set("Location", *url)
	w.WriteHeader(302)
}

// PostURL [POST] /
func PostURL(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Path
	dst := r.URL.Query().Get("dst")

	if src == "" || dst == "" {
		respondJSON(w, 400, nil, map[string]interface{}{
			"result": "error",
			"detail": "require dst",
		})
		return
	}

	log.Print("src:" + src)
	log.Print("dst:" + dst)

	registerRedirectURL(src, dst)
	respondJSON(w, 200, nil, map[string]interface{}{"result": "success"})
}

// panic時リカバーミドルウェア
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				log.Printf("panic: %+v", err)
				respondJSON(w, 500, nil, map[string]interface{}{"Internal Server Error": err})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// HTTPステータスコードを指定してjsonを返す
func respondJSON(w http.ResponseWriter, code int, header http.Header, bodyJSON map[string]interface{}) {
	// ヘッダ設定
	for k, values := range header {
		for _, v := range values {
			log.Printf("%s: %s\n", k, v)
			w.Header().Set(k, v)
		}
	}

	// body設定
	body := []byte{}
	if bodyJSON != nil {
		body, _ = json.Marshal(bodyJSON)
		log.Print(string(body))
	}
	w.WriteHeader(code)
	w.Write(body)
}

func registerRedirectURL(src string, dst string) {
	file, _ := ioutil.ReadFile(filePath)

	urlMap := map[string]string{}
	err := json.Unmarshal(file, &urlMap)
	if err != nil {
		urlMap = map[string]string{}
	}

	urlMap[src] = dst

	updated, _ := json.MarshalIndent(urlMap, "", "  ")

	err = ioutil.WriteFile(filePath, updated, 0644)
	if err != nil {
		log.Print(err)
	}
}

func getRedirectURL(src string) *string {
	file, _ := ioutil.ReadFile(filePath)

	tokens := map[string]string{}
	err := json.Unmarshal(file, &tokens)
	if err != nil {
		tokens = map[string]string{}
	}

	dst, ok := tokens[src]
	if !ok {
		return nil
	}
	return &dst
}
