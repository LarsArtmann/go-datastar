module github.com/larsartmann/go-datastar

go 1.26.5

require (
	github.com/larsartmann/go-error-family v0.10.0
	github.com/larsartmann/go-sse v0.4.0
)

require (
	github.com/larsartmann/go-branded-id v0.5.1 // indirect
	github.com/larsartmann/go-datastar/datastartest v0.0.0-00010101000000-000000000000
	github.com/larsartmann/go-datastar/static v0.0.3
)

replace github.com/larsartmann/go-datastar/datastartest => ./datastartest

replace github.com/larsartmann/go-datastar/static => ./static
