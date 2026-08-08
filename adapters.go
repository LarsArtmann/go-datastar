package datastar

import (
	"context"
	"io"
	"strings"

	errorfamily "github.com/larsartmann/go-error-family"
)

// TemplComponent satisfies the component rendering interface for the [Templ]
// template engine. This separate type ensures compatibility with Templ without
// imposing a dependency on those who prefer a different template engine.
//
// [Templ]: https://templ.guide/
type TemplComponent interface {
	Render(ctx context.Context, w io.Writer) error
}

// ElementsFromTempl renders a [TemplComponent] to HTML and creates an
// [ElementsPatch] from the result.
func ElementsFromTempl(c TemplComponent, opts ...ElementPatchOption) (ElementsPatch, error) { //nolint:erraudit // returns error interface by design — idiomatic Go, consistent with go-sse
	var buf strings.Builder
	if err := c.Render(context.Background(), &buf); err != nil {
		return ElementsPatch{}, errorfamily.Wrapf(err, errorfamily.Orchestration,
			CodeTemplRenderFailed, "render templ component to HTML")
	}

	return NewElementsPatch(buf.String(), opts...), nil
}

// GoStarElementRenderer satisfies the component rendering interface for the
// [GoStar] template engine.
//
// [GoStar]: https://github.com/delaneyj/gostar
type GoStarElementRenderer interface {
	Render(w io.Writer) error
}

// ElementsFromGostar renders a [GoStarElementRenderer] to HTML and creates an
// [ElementsPatch] from the result.
func ElementsFromGostar(
	r GoStarElementRenderer,
	opts ...ElementPatchOption,
) (ElementsPatch, error) { //nolint:erraudit // returns error interface by design — idiomatic Go, consistent with go-sse
	r GoStarElementRenderer,
	opts ...ElementPatchOption,
) (ElementsPatch, error) {
	var buf strings.Builder
	if err := r.Render(&buf); err != nil {
		return ElementsPatch{}, errorfamily.Wrapf(err, errorfamily.Orchestration,
			CodeGostarRenderFailed, "render gostar element to HTML")
	}

	return NewElementsPatch(buf.String(), opts...), nil
}
