////////////////////////////////////////
// server
//
// lots of private functions in net/http/fs.go which mirror what
// server needs to do.
//
////////////////////////////////////////
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

////////////////////////////////////////
// global
////////////////////////////////////////

var n2l *lockUrnDb

////////////////////////////////////////

func server(port int, repos string) {
	pwd, _ := os.Getwd()
	if err := os.Chdir(repos); err != nil {
		log.Fatal(err)
	}
	defer os.Chdir(pwd)

	n2l = &lockUrnDb{}
	updateN2LfromDisk(".")
	dumpN2L(n2l)

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/N2Ls", n2lHandler)
	http.HandleFunc("/N2C", n2cHandler)
	http.HandleFunc("/resource/", resourceHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func dumpN2L(db *lockUrnDb) {
	urns := db.keys()
	for _, urn := range urns {
		fmt.Println(urn)
		urls, _ := db.get(urn)
		for _, url := range urls {
			fmt.Println("  ", url)
		}
	}
}

func updateN2LfromDisk(reposDir string) (err error) {
	err = filepath.Walk(reposDir, walkFn)
	return
}

func walkFn(path string, info os.FileInfo, err error) error {
	parts := strings.Split(path, string(filepath.Separator))
	if len(parts) == 2 {
		if isHashInvalid(parts[1]) {
			return fmt.Errorf("invalid item in repository: %s", path)
		}
		urn := fmt.Sprintf("urn:%s:%s:%s", nis, parts[0], parts[1])
		n2l.append(urn, urlFromRemoteAndResource(&rem, parts[1]))
		return filepath.SkipDir
	}
	return nil
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Amber</h1><p>Coming soon...</p>")
}

func n2lHandler(w http.ResponseWriter, r *http.Request) {
	// http://localhost:8080/N2Ls?urn:x-amber:resource:abc123
	// PREREQ: n2l dictionary is not nil
	log.Printf("%v %v", r.Method, r.RequestURI)

	if r.Method != "GET" {
		http.Error(w, "method not allowed: "+r.Method, http.StatusMethodNotAllowed)
		return
	}

	i := strings.Index(r.RequestURI, "?")
	if i == -1 {
		http.Error(w, "cannot find ?: "+r.RequestURI, http.StatusBadRequest)
		return
	}

	query := r.RequestURI[i+1:]
	parts := strings.Split(query, ":")
	if len(parts) != 4 {
		http.Error(w, "cannot find 3 colons: "+query, http.StatusBadRequest)
		return
	}
	// urn
	if parts[0] != "urn" {
		http.Error(w, "cannot find urn: "+query, http.StatusBadRequest)
		return
	}
	// nid
	if parts[1] != "x-amber" && parts[1] != "amber" {
		http.Error(w, "NID is not amber: "+query+": "+parts[1], http.StatusBadRequest)
		return
	}
	// nss (resource)
	if parts[2] != "resource" {
		http.Error(w, "NSS ought start with resource: "+query+": "+parts[1], http.StatusBadRequest)
		return
	}
	// nss (cHash)
	cHash := parts[3]
	if cHash == "" {
		http.Error(w, "empty resource hash in NSS: "+query, http.StatusBadRequest)
		return
	}

	// look up
	if urls, ok := n2l.get(query); ok {
		w.Header().Set("Content-Type", "text/uri-list; charset=utf-8")

		var response bytes.Buffer
		response.WriteString("# ")
		response.WriteString(query)
		response.WriteString(crlf)
		for _, item := range urls {
			response.WriteString(item)
			response.WriteString(crlf)
		}
		w.Write(response.Bytes())
		// w.WriteHeader(303)
		return
	}
	// OPTIONAL: before failure, resolve URN by peer query
	http.NotFound(w, r)
	return
}

func n2cHandler(w http.ResponseWriter, r *http.Request) {
	// http://localhost:8080/N2C?urn:x-amber:resource:abc123
	log.Printf("%v %v", r.Method, r.RequestURI)

	if r.Method != "GET" {
		http.Error(w, "method not allowed: "+r.Method, http.StatusMethodNotAllowed)
		return
	}

	i := strings.Index(r.RequestURI, "?")
	if i == -1 {
		http.Error(w, "cannot find ?: "+r.RequestURI, http.StatusBadRequest)
		return
	}

	query := r.RequestURI[i+1:]
	parts := strings.Split(query, ":")
	if len(parts) != 4 {
		http.Error(w, "cannot find 3 colons: "+query, http.StatusBadRequest)
		return
	}
	// urn
	if parts[0] != "urn" {
		http.Error(w, "cannot find urn: "+query, http.StatusBadRequest)
		return
	}
	// nid
	if parts[1] != "x-amber" && parts[1] != "amber" {
		http.Error(w, "NID is not amber: "+query+": "+parts[1], http.StatusBadRequest)
		return
	}
	// nss (resource)
	if parts[2] != "resource" {
		http.Error(w, "NSS ought start with resource: "+query+": "+parts[1], http.StatusBadRequest)
		return
	}
	// nss (cHash)
	cHash := parts[3]
	if cHash == "" {
		http.Error(w, "empty resource hash in NSS: "+query, http.StatusBadRequest)
		return
	}

	pathname := fmt.Sprintf("resource/%s/meta", cHash)
	sendFileContents(pathname, w, r)
}

func resourceHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL.Path)
	meta, err := request2metadata(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if meta.uName != "-" {
		if err := verifySignature(meta, r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}
	switch {
	case r.Method == "GET":
		resourceGet(meta, w, r)
	case r.Method == "PUT":
		resourcePut(meta, w, r)
	default:
		http.Error(w, "method not allowed: "+r.Method, http.StatusMethodNotAllowed)
		return
	}
}

func resourceGet(meta metadata, w http.ResponseWriter, r *http.Request) {
	metaPathname := fmt.Sprintf("resource/%s/meta", meta.Chash)
	blob, err := ioutil.ReadFile(metaPathname)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	metaFoo, err := parseUrc(blob)
	w.Header().Set("X-Amber-Encryption", metaFoo.eName)
	w.Header().Set("X-Amber-Hash", metaFoo.hName)
	sendFileContents(meta.bpathname, w, r)
}

func resourcePut(meta metadata, w http.ResponseWriter, r *http.Request) {
	if meta.hName == "-" {
		err := fmt.Errorf("hash name cannot be '-': %#v", meta)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if meta.size == "-" {
		err := fmt.Errorf("size cannot be '-': %#v", meta)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = checkHash(meta.hName, bytes, meta.Chash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = writeFile(meta.bpathname, bytes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metablob := fmt.Sprintf("Content-Length: %d\r\n"+
		"X-Amber-Hash: %v\r\n"+
		"X-Amber-Encryption: %v\r\n",
		len(bytes), meta.hName, meta.eName)
	if err = writeFile(meta.mpathname, []byte(metablob)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urn := fmt.Sprintf("urn:%s:resource:%s", nis, meta.Chash)
	n2l.append(urn, urlFromRemoteAndResource(&rem, meta.Chash))
	fmt.Fprintf(w, "<p>%v bytes written to %v</p>", len(bytes), urn)
}

func sendFileContents(pathname string, w http.ResponseWriter, r *http.Request) {
	fi, err := os.Stat(pathname)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if fi.IsDir() {
		err := fmt.Errorf("found dir instead of file: %s", pathname)
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, err := ioutil.ReadFile(pathname)
	if err != nil {
		err := fmt.Errorf("error reading file: %s: %s", pathname, err)
		log.Printf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
	return
}

func verifySignature(meta metadata, r *http.Request) (err error) {
	// How do you check credentials when you put a document?
	// Where are the credentials given?
	// Use a custom header key, with a value set to a signed copy of the body?
	// var sig []byte
	// sig = header[X-Amber-Body-Digest]
	// if sig == "" {
	//	return
	// }
	// var pub *PublicKey // derived from user string (which may cause URLs to become too long)
	// hashed := computeHash(req.hName, req.Chash)
	// if err = rsa.VerifyPKCS1v15(pub, req.hName, hashed, sig); err != nil {
	//	return fmt.Errorf("unauthorized")
	// }
	return fmt.Errorf("unauthorized")
}

////////////////////////////////////////
