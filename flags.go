package soyweb

import (
	"github.com/soyart/ssg-go"
)

// FlagsV2 represents CLI arguments that could modify soyweb behavior, such as skipping stages
// and minifying content of certain file extensions.
type FlagsV2 struct {
	NoCleanup       bool `arg:"--no-cleanup" help:"Skip cleanup stage"`
	NoCopy          bool `arg:"--no-copy" help:"Skip scopy stage"`
	NoBuild         bool `arg:"--no-build" help:"Skip build stage"`
	NoReplace       bool `arg:"--no-replace" help:"Do not do text replacements defined in manifest"`
	NoGenerateIndex bool `arg:"--no-gen-index" help:"Do not generate indexes on _index.soyweb"`

	MinifyHTMLGenerate bool `arg:"--min-html" help:"Minify converted HTML outputs"`
	MinifyHTMLCopy     bool `arg:"--min-html-copy" help:"Minify all copied HTML"`
	MinifyCSS          bool `arg:"--min-css" help:"Minify CSS files"`
	MinifyJs           bool `arg:"--min-js" help:"Minify Javascript files"`
	MinifyJson         bool `arg:"--min-json" help:"Minify JSON files"`
}

type FlagsNoMinify struct {
	NoMinifyHTMLGenerate bool `arg:"--no-min-html,env:NO_MIN_HTML" help:"Do not minify converted HTML outputs"`
	NoMinifyHTMLCopy     bool `arg:"--no-min-html-copy,env:NO_MIN_HTML_COPY" help:"Do not minify all copied HTML"`
	NoMinifyCSS          bool `arg:"--no-min-css,env:NO_MIN_CSS" help:"Do not minify CSS files"`
	NoMinifyJs           bool `arg:"--no-min-js,env:NO_MIN_JS" help:"Do not minify Javascript files"`
	NoMinifyJson         bool `arg:"--no-min-json,env:NO_MIN_JSON" help:"Do not minify JSON files"`
}

func (f FlagsV2) Stage() Stage {
	s := StageAll
	if f.NoCleanup {
		s.Skip(StageCleanUp)
	}
	if f.NoCopy {
		s.Skip(StageCopy)
	}
	if f.NoBuild {
		s.Skip(StageBuild)
	}
	return s
}

func (f FlagsV2) Hooks() []ssg.Hook {
	return filterNilHooks(
		f.hookMinify(),
	)
}

func (f FlagsV2) hookMinify() ssg.Hook {
	m := make(map[string]MinifyFn)
	if f.MinifyHTMLCopy {
		m[".html"] = MinifyHTML
	}
	if f.MinifyCSS {
		m[".css"] = MinifyCSS
	}
	if f.MinifyJs {
		m[".js"] = MinifyJS
	}
	if f.MinifyJson {
		m[".json"] = MinifyJSON
	}
	return HookMinify(m)
}

func (f FlagsNoMinify) Flags() FlagsV2 {
	return FlagsV2{
		MinifyHTMLGenerate: !f.NoMinifyHTMLGenerate,
		MinifyHTMLCopy:     !f.NoMinifyHTMLCopy,
		MinifyCSS:          !f.NoMinifyCSS,
		MinifyJs:           !f.NoMinifyJs,
		MinifyJson:         !f.NoMinifyJson,
	}
}

func (f FlagsNoMinify) Skip(ext string) bool {
	switch ext {
	case ExtHTML:
		if f.NoMinifyHTMLGenerate {
			return true
		}
		if f.NoMinifyHTMLCopy {
			return true
		}
	case ExtCSS:
		if f.NoMinifyCSS {
			return true
		}
	case ExtJS:
		if f.NoMinifyJs {
			return true
		}
	case ExtJSON:
		if f.NoMinifyJson {
			return true
		}

	default:
		// Skip unknown file extension and media type
		return true
	}

	// Do not skip this extension
	return false
}
