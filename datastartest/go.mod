module github.com/larsartmann/go-datastar/datastartest

go 1.26.6

require (
	github.com/larsartmann/go-datastar v0.2.0
	github.com/larsartmann/go-error-family v0.10.0
	github.com/larsartmann/go-sse v0.5.1
)

require (
	github.com/larsartmann/go-branded-id v0.5.1 // indirect
	github.com/larsartmann/go-datastar/static v0.2.0 // indirect
)

replace github.com/larsartmann/go-datastar => ..

replace github.com/larsartmann/go-datastar/static => ../static
