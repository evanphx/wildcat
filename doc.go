/*

Wildcat provides a HTTP parser that performs zero allocations, in scans a buffer
and tracks slices into the buffer to provide information about the request.

As a result, it is also quite fast. `go test -bench .` to verify for yourself.

A couple of cases where it wil perform an allocation:

1. When the parser is created, it makes enough space to track 10 headers. If
there end up being more than 10, this is resized up to 20, then 30, etc.

This can be overriden by using NewSizedHTTPParser to pass in the initial size
to use, eliminated allocations for your use case (you could set this to 1000
for instance).

2. If a mime-style multiline header is encountered, wildcat will make a new
buffer to contain the concatination of values.

NOTE: FindHeader only returns the first header that matches the requested name.
If a request contains multiple values for the same header, use FindAllHeaders.

*/
package wildcat
