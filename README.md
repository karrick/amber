* blob
* tree
** maybe do not need to store full urn of children, but just ehash
* commit
* refs points to commit, not object in DAG
** maybe this has urn associated with it
* HEAD is special ref that points to actual ref, also not object in DAG
* tag points to commit, is object in DAG, and has message and GPG signature

# compression of contents
## could be dynamic, as decided upon the client and decompressed when unpacking it.
## whatever parent object a blob has would determine the compression used.
## already compressed object would be stored as is, and the compression scheme marked none.

# similar thing for determining the encryption algorithm. the parent object would know:

# In other words, the parent object merely described how to reify the child object.

* ehash to locate and check downloaded bytes
* encryption method
* phash to decrypt and check decrypted & decompressed bytes
* compression method
* object class, type (file, directory, symlink)
* reified name
* mode

////////////////////////////////////////
// <uri> ::= <urn> | <url>
//
// I'm looking for the following resource...

// GET /uri-res/<service>?<uri>  HTTP/1.1
// GET /uri-res/N2L?urn:foo:12345-54321 HTTP/1.1

// N2L() lookup where urn lives, success: 200, Location: url; failure: 404
// N2Ls() lookup where urn lives, success: 200, returns zero or more urls; failure: 404
// N2R() lookup where urn lives, success: return resource; 

// UH-OH!  How do I specify user in URN? When given URN, must query N2Ls, but how does it know what to set X-Amber-User to?

// URL:	http://hostport/resource/abc123
// FS:	     repository/resource/abc123/users/-
//                               abc123/meta (relevant headers)

// URL:	http://hostport/account/abc123
// FS:	     repository/account/abc123

// func request2pathname(r *http.Request) (pathname string, err error) {
// 	switch {
// 	case !strings.HasPrefix(r.URL.Path, "/resource/"):
// 		fallthrough
// 	case isHashInvalid(r.URL.Path[len("/resource/"):]):
// 		err = fmt.Errorf("illegal url: %s", r.URL.Path)
// 		return
// 	}

// 	user, err := mustLookupHeader(r.Header, "X-Amber-User")
// 	if err != nil {
// 		user = "-" ; err = nil
// 	}
// 	if user != "-" && isHashInvalid(user) {
// 		err = fmt.Errorf("invalid user: %s", user)
// 		return
// 	}

// 	pathname = fmt.Sprintf("%s/users/%s", r.URL.Path[1:], user)
// 	return
// }

// TO PREVENT DOS, by default server will not allow upload directly to
// - user, but require upload to valid user space first, then hardlink
// to - user provided cred given for account that has that resource.
// SOME SERVERS may be configured to accept direct upload to - user.

// HEAD + GET Headers:
//
// Content-Length: 1234
// X-Amber-Encryption: aes256
// X-Amber-Hash: sha1
// X-Amber-Mode: 644

