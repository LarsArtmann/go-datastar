module github.com/larsartmann/go-datastar

go 1.26.7

require (
	github.com/larsartmann/go-error-family v0.10.0
	github.com/larsartmann/go-sse v0.6.0
)

require golang.org/x/mod v0.40.0

require (
	github.com/larsartmann/go-branded-id v0.5.1 // indirect
	github.com/larsartmann/go-datastar/static v0.3.0
)

replace github.com/larsartmann/go-datastar/static => ./static
