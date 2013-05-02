`gotogether` - Resource Compiler for Go
=================================

`gotogether`  a directory of resource into a Go source file so you can still
deploy a single executable as a web server with all the CSS, image files, JS ...
included.

This is a fork of tebeka's nrsc: http://bitbucket.org/tebeka/nrsc

Installing
==========
::

    go get github.com/carbocation/gotogether

Also grab the `gotogether` script from here_

And you'll need `zip` somewhere in your path.

.. _here: http://bit.ly/gotogether-script

Invocation
==========
::

    go build
    gotogether <executable> <resource dir> [zip options]


API
===
The `gotogether` package has the following interface

`gotogether.Handle(prefix string)`
    This will register with the `net/http` module to handle all paths starting with prefix. 

    When a request is handled, `prefix` is stripped and then a resource is
    located and served.

    Resource that are not found will cause an HTTP 404 response.
    

`gotogether.Get(path string) Resource`
    Will return a resource interface (or `nil` if not found) (see resource interface below).
    This allows you more control on how to serve.


`LoadTemplates(t *template.Template, filenames ...string) (*template.Template, error)`
    Will load named templates from resources. If the argument "t" is `nil`, it is
    created from the first resource.

Resource Interface
------------------

`func Open() (io.Reader, error)`
    Returns a reader to resource data

`func Size() int64`
    Returns resource size (to be used with `Content-Length` HTTP header).

`func ModTime() time.Time`
    Returns modification time (to be used with `Last-Modified` HTTP header).


Example Code
------------
::

    package main

    import (
            "fmt"
            "net/http"
            "os"

            "github.com/carbocation/gotogether"
    )

    func indexHandler(w http.ResponseWriter, req *http.Request) {
            fmt.Fprintf(w, "Hello World\n")
    }

    func main() {
            gotogether.Handle("/static/")
            http.HandleFunc("/", indexHandler)
            if err := http.ListenAndServe(":8080", nil); err != nil {
                    fmt.Fprintf(os.Stderr, "error: %s\n", err)
                    os.Exit(1)
            }
    }

Contact
=======
https://github.com/carbocation/gotogether
    
License
=======
MIT (see `LICENSE.txt`_)

.. _`LICENSE.txt`: https://github.com/carbocation/gotogether/src/tip/LICENSE.txt
