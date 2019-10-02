package repeater

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/kabukky/httpscerts"
)

// Server

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	strId := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(strId, 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	req, err := getRequest(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := http.Client{}
	bodyReader := bytes.NewReader(req.body)
	request, err := http.NewRequest(req.method, req.uri, bodyReader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	header := make(map[string]string, 100)
	err = json.Unmarshal(req.header, &header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for k, v := range header {
		request.Header.Add(k, v)
	}

	resp, err := client.Do(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy header
	w.WriteHeader(resp.StatusCode)
	copyHeader(w.Header(), resp.Header)

	// Copy body
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

func startHttp() {
	// Generate certificates
	err := httpscerts.Check(certPath, keyPath)
	if err != nil {
		err = httpscerts.Generate(certPath, keyPath, host+":"+port)
		if err != nil {
			log.Fatalln("Impossible to generate https certificate.")
		}
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(mainHandler),
		// Disable HTTP/2 because it incompatible with Hijacker
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	if protocol == "http" {
		log.Println("Starting HTTP repeater server at " + host + ":" + port)
		log.Fatalln(server.ListenAndServe())
	} else if protocol == "https" {
		log.Println("Starting HTTPS repeater server at " + host + ":" + port)
		log.Fatalln(server.ListenAndServeTLS(certPath, keyPath))
	} else {
		log.Fatalln("Unknown protocol!")
	}
}
