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

func server(rem remote, repos string) {
	pwd, _ := os.Getwd()
	if err := os.Chdir(repos); err != nil {
		log.Fatal(err)
	}
	defer os.Chdir(pwd)

	log.Print("inventorying existing resources")
	n2l = &lockUrnDb{}
	updateN2LfromDisk(".")
	dumpN2L(n2l)

	log.Print("setting up web service")
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/N2Ls", n2lHandler)
	http.HandleFunc("/N2C", n2cHandler)
	http.HandleFunc("/resource/", resourceHandler)
	hostport := fmt.Sprintf("%s:%d", rem.hostname, rem.port)
	log.Printf("listening for connections: %s", hostport)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", rem.port), nil))
}

func dumpN2L(db *lockUrnDb) {
	urns := db.keys()
	for _, urn := range urns {
		if debug {
			fmt.Println(urn)
		}
		urls, _ := db.get(urn)
		for _, url := range urls {
			if debug {
				fmt.Println("  ", url)
			}
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
	log.Printf("%v %v", r.Method, r.RequestURI)

	if r.Method != "GET" {
		err := fmt.Errorf("method not allowed: %s", r.Method)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	i := strings.Index(r.RequestURI, "?")
	if i == -1 {
		err := fmt.Errorf("cannot find ?: %s", r.RequestURI)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := r.RequestURI[i+1:]
	parts := strings.Split(query, ":")
	if len(parts) != 4 {
		err := fmt.Errorf("cannot find 3 colons: %s", query)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// urn
	if parts[0] != "urn" {
		err := fmt.Errorf("cannot find urn: %s", query)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// nid
	if parts[1] != "x-amber" && parts[1] != "amber" {
		err := fmt.Errorf("NIS is not amber: %s:%s", query, parts[1])
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// nss (resource)
	if parts[2] != "resource" {
		err := fmt.Errorf("NSS ought start with resource: %s:%s", query, parts[1])
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// nss (cHash)
	cHash := parts[3]
	if cHash == "" {
		err := fmt.Errorf("empty resource hash in NSS: %s", query)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		w.WriteHeader(303)
		w.Write(response.Bytes())
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
		err := fmt.Errorf("method not allowed: %s", r.Method)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	i := strings.Index(r.RequestURI, "?")
	if i == -1 {
		err := fmt.Errorf("cannot find ?: %s", r.RequestURI)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := r.RequestURI[i+1:]
	parts := strings.Split(query, ":")
	if len(parts) != 4 {
		err := fmt.Errorf("cannot find 3 colons: %s", query)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// urn
	if parts[0] != "urn" {
		err := fmt.Errorf("cannot find urn: %s", query)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// nid
	if parts[1] != "x-amber" && parts[1] != "amber" {
		err := fmt.Errorf("NID is not amber: %s:%s", query, parts[1])
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// nss (resource)
	if parts[2] != "resource" {
		err := fmt.Errorf("NSS ought start with resource: %s:%s", query, parts[1])
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// nss (cHash)
	cHash := parts[3]
	if cHash == "" {
		err := fmt.Errorf("empty resource hash in NSS: %s", query)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pathname := fmt.Sprintf("resource/%s/meta", cHash)
	sendFileContents(pathname, w, r)
}

func resourceHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL.Path)
	meta, err := request2metadata(r)
	if err != nil {
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if meta.uName != "-" {
		if err := verifySignature(meta, r); err != nil {
			if debug {
				log.Print(err)
			}
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
		err := fmt.Errorf("method not allowed: %s", r.Method)
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}
}

func resourceGet(meta metadata, w http.ResponseWriter, r *http.Request) {
	metaPathname := fmt.Sprintf("resource/%s/meta", meta.Chash)
	blob, err := ioutil.ReadFile(metaPathname)
	if err != nil {
		if os.IsNotExist(err) {
			if debug {
				log.Print(err)
			}
			http.NotFound(w, r)
			return
		}
		if debug {
			log.Print(err)
		}
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
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err = checkHash(meta.hName, bytes, meta.Chash); err != nil {
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = writeFileNoOverwrite(meta.bpathname, bytes); err != nil {
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metablob := fmt.Sprintf("Content-Length: %d\r\n"+
		"X-Amber-Hash: %v\r\n"+
		"X-Amber-Encryption: %v\r\n",
		len(bytes), meta.hName, meta.eName)
	if err = writeFileNoOverwrite(meta.mpathname, []byte(metablob)); err != nil {
		if debug {
			log.Print(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urn := fmt.Sprintf("urn:%s:resource:%s", nis, meta.Chash)
	n2l.append(urn, urlFromRemoteAndResource(&rem, meta.Chash))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%v bytes written to %v", len(bytes), urn)
}

func sendFileContents(pathname string, w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadFile(pathname)
	if err != nil {
		if debug {
			log.Print(err)
		}
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Length", fmt.Sprint(len(bytes)))
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
