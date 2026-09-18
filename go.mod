module github.com/larsartmann/go-datastar

go 1.26.7

require (
	github.com/larsartmann/go-datastar/static v0.6.0
	github.com/larsartmann/go-error-family v0.10.1
	github.com/larsartmann/go-sse v0.6.0
	golang.org/x/mod v0.41.0
)

require github.com/larsartmann/go-branded-id v0.6.0 // indirect

replace github.com/larsartmann/go-datastar/static => ./static
