module github.com/larsartmann/go-datastar/broadcast

go 1.27.1

require (
	github.com/larsartmann/go-datastar v0.6.0
	github.com/larsartmann/go-sse v0.6.0
)

require (
	github.com/larsartmann/go-branded-id v0.6.0 // indirect
	github.com/larsartmann/go-datastar/static v0.6.0 // indirect
	github.com/larsartmann/go-error-family v0.10.1 // indirect
)

replace github.com/larsartmann/go-datastar => ..

replace github.com/larsartmann/go-datastar/static => ../static
