module github.com/larsartmann/go-datastar

go 1.27

require (
	github.com/larsartmann/go-datastar/static v0.7.0
	github.com/larsartmann/go-error-family v0.12.0
	github.com/larsartmann/go-sse v0.7.0
	golang.org/x/mod v0.43.0
)

require github.com/larsartmann/go-branded-id v0.7.1 // indirect

replace github.com/larsartmann/go-datastar/static => ./static
