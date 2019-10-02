package sproxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/kabukky/httpscerts"
)

func makeRequest(method, uri string, header http.Header, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil
	}
	request.Header = header
	return request
}

func saveRequest(req *http.Request) *http.Request {
	headerMap := make(map[string]string, 100)
	for k, vv := range req.Header {
		for _, v := range vv {
			headerMap[k] = v
		}
	}
	header, err := json.Marshal(headerMap)
	if err != nil {
		log.Println("impossible marshall header " + err.Error())
		return req
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("impossible to copy body: " + err.Error())
		return req
	}

	err = req.Body.Close()
	if err != nil {
		log.Println("impossible to close body: " + err.Error())
		return req
	}

	err = insertRequest(request{
		uri:    req.RequestURI,
		method: req.Method,
		header: header,
		body:   body,
	})
	if err != nil {
		log.Println("impossible to save in db " + err.Error())
	}

	return makeRequest(req.Method, req.RequestURI, req.Header, bytes.NewReader(body))
}

// HTTP

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	req = saveRequest(req)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
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

// HTTPS

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer func() {
		if err := destination.Close(); err != nil {
			log.Println(err)
		}
	}()
	defer func() {
		if err := source.Close(); err != nil {
			log.Println(err)
		}
	}()

	log.Println("begin reading")
	_, err := io.Copy(destination, source)
	if err != nil {
		log.Println(err)
	}
	log.Println("end reading")
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Println("no connect")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("no hijacker")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Println("no hijack")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	// w.WriteHeader(http.StatusOK)
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

// Server

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		log.Println("here need be https")
		// handleTunneling(w, r)
	} else {
		handleHTTP(w, r)
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
		log.Println("Starting HTTP proxy server at " + host + ":" + port)
		log.Fatalln(server.ListenAndServe())
	} else if protocol == "https" {
		log.Println("Starting HTTPS proxy server at " + host + ":" + port)
		log.Fatalln(server.ListenAndServeTLS(certPath, keyPath))
	} else {
		log.Fatalln("Unknown protocol!")
	}
}
