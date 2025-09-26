package soyweb

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/soyart/ssg-go"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
)

type (
	MinifyFn func(data []byte) ([]byte, error)
)

const (
	MediaTypeHTML = "text/html"
	MediaTypeCSS  = "style/css"
	MediaTypeJS   = "text/javascript"
	MediaTypeJSON = "application/json"

	ExtHTML = ".html"
	ExtCSS  = ".css"
	ExtJS   = ".js"
	ExtJSON = ".json"
)

var m = minify.New()

func init() {
	m.Add(MediaTypeHTML, &html.Minifier{
		// Default values are shown to be more conspicuous
		KeepComments:            false,
		KeepConditionalComments: false,
		KeepSpecialComments:     false,
		KeepDefaultAttrVals:     false,
		KeepDocumentTags:        true,
		KeepEndTags:             true,
		KeepQuotes:              false,
		KeepWhitespace:          false,
		TemplateDelims:          [2]string{},
	})
	m.AddFunc(MediaTypeCSS, css.Minify)
	m.AddFunc(MediaTypeJS, js.Minify)
	m.AddFunc(MediaTypeJSON, json.Minify)
}

func MinifyHTML(og []byte) ([]byte, error) { return minifyMedia(og, MediaTypeHTML) }
func MinifyCSS(og []byte) ([]byte, error)  { return minifyMedia(og, MediaTypeCSS) }
func MinifyJS(og []byte) ([]byte, error)   { return minifyMedia(og, MediaTypeJS) }
func MinifyJSON(og []byte) ([]byte, error) { return minifyMedia(og, MediaTypeJSON) }

func MinifyAll(path string, data []byte) ([]byte, error) {
	fn, err := ExtToFn(filepath.Ext(path))
	if err != nil {
		return data, nil
	}
	out, err := fn(data)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func MinifyFile(path string) ([]byte, error) {
	original, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer original.Close()

	mediaType, err := ExtToMediaType(filepath.Ext(path))
	if err != nil {
		return nil, err
	}

	min := bytes.NewBuffer(nil)
	err = m.Minify(mediaType, min, original)
	if err != nil {
		return nil, err
	}

	return min.Bytes(), nil
}

func ExtToMediaType(ext string) (string, error) {
	switch ext {
	case ExtHTML:
		return MediaTypeHTML, nil
	case ExtCSS:
		return MediaTypeCSS, nil
	case ExtJS:
		return MediaTypeJS, nil
	case ExtJSON:
		return MediaTypeJSON, nil
	}
	return "", fmt.Errorf("'%s': %w", ext, ErrWebFormatNotSupported)
}

func ExtToFn(ext string) (func([]byte) ([]byte, error), error) {
	switch ext {
	case ExtHTML:
		return MinifyHTML, nil
	case ExtCSS:
		return MinifyCSS, nil
	case ExtJS:
		return MinifyJS, nil
	case ExtJSON:
		return MinifyJSON, nil
	}
	return nil, fmt.Errorf("'%s': %w", ext, ErrWebFormatNotSupported)
}

func HookMinifyDefault(mediaTypes ssg.Set) ssg.Hook {
	m := make(map[string]MinifyFn)
	if mediaTypes.Contains(MediaTypeHTML) {
		m[ExtHTML] = MinifyHTML
	}
	if mediaTypes.Contains(MediaTypeJS) {
		m[ExtJS] = MinifyJS
	}
	if mediaTypes.Contains(MediaTypeJSON) {
		m[ExtJSON] = MinifyJSON
	}
	if mediaTypes.Contains(MediaTypeCSS) {
		m[ExtCSS] = MinifyCSS
	}
	return HookMinify(m)
}

func HookMinify(m map[string]MinifyFn) ssg.Hook {
	if len(m) == 0 {
		return nil
	}
	return func(path string, data []byte) ([]byte, error) {
		ext := filepath.Ext(path)
		f, ok := m[ext]
		if !ok {
			return data, nil
		}
		b, err := f(data)
		if err != nil {
			return nil, fmt.Errorf("error from minifier for '%s'", ext)
		}
		return b, nil
	}
}

func minifyMedia(original []byte, mediaType string) ([]byte, error) {
	min := bytes.NewBuffer(nil)
	err := m.Minify(mediaType, min, bytes.NewBuffer(original))
	if err != nil {
		return nil, err
	}
	return min.Bytes(), nil
}
