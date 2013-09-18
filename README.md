////////////////////////////////////////
// <uri> ::= <urn> | <url>
//
// I'm looking for the following resource...

// GET /uri-res/<service>?<uri>  HTTP/1.1
// GET /uri-res/N2L?urn:foo:12345-54321 HTTP/1.1

// N2L() lookup where urn lives, success: 200, Location: url; failure: 404
// N2Ls() lookup where urn lives, success: 200, returns zero or more urls; failure: 404
// N2R() lookup where urn lives, success: return resource; 

