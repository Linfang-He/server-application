## A simple server application that accepts requests using ListenAndServe

### Basic data structures for http communication
```
type Server struct {
	// Addr ("host:port") : specifies the TCP address of the server
	Addr string
	// DocRoot the root folder under which clients can potentially look up information.
	// Anything outside this should be "out-of-bounds"
	DocRoot string
}

type Request struct {
	Method string // e.g. "GET"
}

type Response struct {
	StatusCode int    // e.g. 200 / 405
	Proto string	  // HTTP/1.1
	FilePath string   // For this application, we will hard-code this to whatever contents are available in "hello-world.txt"
}

```

```
func (s *Server) ListenAndServe() error {}
```
[The ListenAndServe from net/http](https://pkg.go.dev/net/http#ListenAndServe)

### Implementing ListenAndServe

First, we validate whether the server is set up properly. In this example, we just need to see whether
the doc root is correctly set up or not.

os.Stat : returns the FileInfo structure describing file <br/>
[os.IsNotExist](https://pkg.go.dev/os#IsNotExist)
```go
// Validating the doc root of the server
fi, err := os.Stat(s.DocRoot)

if os.IsNotExist(err) {
    return err
}

if !fi.IsDir() {
    return fmt.Errorf("doc root %q is not a directory", s.DocRoot)
}

return nil
```


```go
// server should now start to listen on the configured address
ln, err := net.Listen("tcp", s.Addr)
if err != nil {
    return err
}
fmt.Println("Listening on", ln.Addr())
```

start accepting connections
```go
// accept connections forever
for {
    conn, err := ln.Accept()
    if err != nil {
    continue
    }
    fmt.Println("accepted connection", conn.RemoteAddr())
    go s.HandleConnection(conn)
}
```

### Implementing HandleConnection

Set timeout for every Read operation
```go
conn.SetReadDeadline(time.Now().Add(5 * time.Second))
```

Read next request from the client
```go
req, err := ReadRequest(br)
```

### Implementing ReadRequest

Read the start line of the Request
We'll use the handy method `func ReadLine(br *bufio.Reader) (string, error)` from util.go
for this. It's the same implementation that's given to you in proj3.

Parse the request line read and do relevant checks
Example : "GET /foo HTTP/1.1" --> ["GET", "/foo HTTP/1.1"] <br/>
[strings.SplitN](https://pkg.go.dev/strings#SplitN)
```go
fields := strings.SplitN(line, " ", 2)
if len(fields) != 2 {
return "", fmt.Errorf("could not parse the request line, got fields %v", fields)
}
return fields[0], nil
```

Read the remaining lines from the request until we get an EOF.


### Back to HandleConnection

Check for the error that ReadRequest returns. It could be the case that
- The client has closed the connection `errors.Is(err, io.EOF)`
- Timeout has happened `err.Timeout()`
- The request from the client is invalid in which case we call `HandleBadRequest`

If all goes well, and we get a proper `Request` object from `ReadRequest`,
we call `HandleGoodRequest`

Here, this we'll not close the connection and handle as many requests for this
connection and pass on the responsibility of maintaining sanity of waiting to the timeout
mechanism.

### Using curl and breakpoint-based debugging of the application

1. Successful GET request
```
╰─ curl -v localhost:8090                                                                                                                                                                                                  ─╯
*   Trying ::1:8090...
* Connection failed
* connect to ::1 port 8090 failed: Connection refused
*   Trying 127.0.0.1:8090...
* Connected to localhost (127.0.0.1) port 8090 (#0)
> GET / HTTP/1.1
> Host: localhost:8090
> User-Agent: curl/7.71.1
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Hello, world!
```

2. Unacceptable HTTP verb
```
╰─ curl -v -XPOST localhost:8090                                                                                                                                                                                           ─╯
*   Trying ::1:8090...
* Connection failed
* connect to ::1 port 8090 failed: Connection refused
*   Trying 127.0.0.1:8090...
* Connected to localhost (127.0.0.1) port 8090 (#0)
> POST / HTTP/1.1
> Host: localhost:8090
> User-Agent: curl/7.71.1
> Accept: */*
>
* Mark bundle as not supporting multiuse
  < HTTP/1.1 405 Method Not Allowed
 ```

3. Invalid HTTP verb
```
╰─ curl -v -XPOS localhost:8090                                                                                                                                                                                            ─╯
*   Trying ::1:8090...
* Connection failed
* connect to ::1 port 8090 failed: Connection refused
*   Trying 127.0.0.1:8090...
* Connected to localhost (127.0.0.1) port 8090 (#0)
> POS / HTTP/1.1
> Host: localhost:8090
> User-Agent: curl/7.71.1
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 405 Method Not Allowed 
```

4. Timeout initiated by the server

5. Connection closed by the `curl` client, io.EOF on the server



### Unit test framework for the server application.

Unit testing, as the name suggests, is usually designed to test basic units of the code
that one writes. It's almost always a good idea to divide your application into
various components/services that deal with different "high-level" ideas. Such as in project3,
we have separate files for response, request and server. These give you a way to test small
pieces of the code and build a larger service with reliability.

The smallest testable parts are called 'units' which are tested independently of all other
units in the system.
Go provides us with a standard library package 'testing' to write automated unit tests.
The way you test in go is by using a single command `go test` and some conventions to write tests:
- The `go test` command looks for the functions of the following signature func **TestXxx(\*testing.T)**
- Xxx can be any alphanumeric string that starts with an uppercase letter
- `go test` will run test files that have a suffix `_test.go`.

`testing.T` ([package](https://pkg.go.dev/testing#pkg-index))
```go
type T struct {
	common
	isParallel bool
	isEnvSet   bool
	context    *testContext // For running tests and subtests.
}
```

`go test` has several, very interesting functionalities, associated with flags to this call.
To check them out run `go help testflag`. One such flag is `-cover` which provides you
with a code coverage feature using your unit tests.

`server_test.go` <br/>
- `SIMPLE GET`
```
request := "GET /index.html HTTP/1.1\r\n" +
           "Host: test\r\n" +
           "\r\n"
```

- `SIMPLE POST`
```
request := "POST /index.html HTTP/1.1\r\n" +
            "Host: test\r\n" +
            "\r\n"
```

### go test usage from the command line

```
go test
``` 
v/s
```
go test -v
```

For various commands checkout `go help test`

```
go test -v -cover
```

Running an individual test
```
go test -v -run TestHandleConnection_Simple_GET
```

#### Table tests

Table-driven tests:
Each table entry is a complete test case with inputs and expected results, and sometimes with
additional information such as a test name to make the test output easily readable.

- The actual test simply iterates through all table entries and for each entry performs the necessary tests.
- The test code is written once and amortized over all table entries, so it makes sense to write a careful test with good error messages
- In most cases the table is a slice of anonymous structs, which allows the table to be written in a compact form.


Running an individual table test
```
go test -v -run TestReadMultipleRequests/GoodBad
```

server_tableTests_test.go
```
var tests = []struct {
		name    string
		reqText string
		reqWant *Request
	}{
		{
			"Simple GET",
			"GET /index.html HTTP/1.1\r\n" +
				"Host: test\r\n" +
				"\r\n",
			&Request{
				Method: "GET",
			},
		},
	}
```

```
var tests = []struct {
		name string
		req  string
	}{
		{
			"Basic",
			"This is a bad request\r\n",
		},
		{
			"Empty",
			"\r\n",
		},
		{
			"InvalidHTTPVerb",
			"GETT /index.html HTTP/1.1\r\nHost: test\r\n\r\n",
		},
		{
			"NotSupportedHTTPVerb",
			"POST /index.html HTTP/1.0\r\nHost: test\r\n\r\n",
		},
	}
```
Running multiple tests as separate go-routines
[t.Run](https://pkg.go.dev/testing#T.Run)


#### Parallel tests

[Some interesting and lesser known things about go test](https://splice.com/blog/lesser-known-features-go-test/)

The testing package in Go also lets us parallelize tests so that a test suite
can run faster. This is accomplished by the use of `testing.Parallel `.

```go
t.Run(tt.name, func(t *testing.T) {
    t.Parallel()
}
```

We've introduced a sleep in server_parallelTests_tests to show how parallel test execution
can speed up the test suite.

The number of tests run simultaneously in parallel is the current value of `GOMAXPROCS` by default.
The following will run 4 tests in parallel.
`go test -parallel 4`

Try running the following with `t.Parallel()` in the test code commented out.
```
go test -v -run TestReadBadRequest_parallel
```
Now, run the same command again with t.Parallel() and notice the time difference between the two runs.


PS : Link to the Goland IDE is [here](https://www.jetbrains.com/go/). Check it out if you want, it's pretty cool!
Also, if you're interested see [Breakpoints in VSCode for go](https://github.com/golang/vscode-go/blob/master/docs/debugging.md)
