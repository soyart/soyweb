package soyweb_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/soyart/soyweb"
	"github.com/soyart/ssg-go"
)

func TestGenerateIndex(t *testing.T) {
	src := "./testdata/myblog/src"
	dst := "./testdata/myblog/dst-test-generate-index"
	title := "TestTitle"
	url := "https://my.blog.testgenindex"

	err := os.RemoveAll(dst)
	if err != nil {
		panic(err)
	}

	defaultTitleHTML := fmt.Sprintf("<title>%s</title>", title)
	contains := map[string][]string{
		"_index.soyweb": {
			`<title>My blog (title tag)</title>`,
			`<li><p><a href="/2023/">2023</a></p></li>`,
			`<li><p><a href="/2022/">twenty and twenty-two (from-tag)</a></p></li>`,
			`<li><p><a href="/testdir/">testdir Title from tag</a></p></li>`,
		},
		"2022/_index.soyweb": {
			`<title>twenty and twenty-two (from-tag)</title>`,
			`<li><p><a href="/2022/bar/">bar</a></p></li>`,
			`<li><p><a href="/2022/foo.html">Foo</a></p></li>`,
		},
		"2023/_index.soyweb": {
			defaultTitleHTML,
			// Because _index.soyweb was empty, so default index header is used.
			`Index of 2023`,
			`<li><p><a href="/2023/baz.html">Bazketball</a></p></li>`,
			`<li><p><a href="/2023/recurse/">Recurse Index</a></p></li>`,
			`<li><p><a href="/2023/lol/">LOLOLOL</a></p></li>`,
		},
		"2023/recurse/_index.soyweb": {
			defaultTitleHTML,
			`<h1 id="recurse-index">Recurse Index</h1>`,
			"<li><p><a href=\"/2023/recurse/a1/\">A1 from Marker H1</a></p></li>",
			"<li><p><a href=\"/2023/recurse/r1/\">Recursive 1</a></p></li>",
			"<li><p><a href=\"/2023/recurse/r2/\">Recursive 2</a></p></li>",
			"<li><p><a href=\"/2023/recurse/r3/\">Recursive 3 from tag</a></p></li>",
		},
		"testdir/_index.soyweb": {
			`<title>testdir Title from tag</title>`,
			`<li><p><a href="/testdir/dir1/">Dir-1-Title-From-Tag</a></p></li>`,
			`<li><p><a href="/testdir/dir2/">Dir-2</a></p></li>`,
			`<li><p><a href="/testdir/testprefer/">testprefer</a></p></li>`,
		},
		"testdir/testprefer/_index.soyweb": {
			defaultTitleHTML,
			// Use dirname because existing index.html is preferred
			`<li><p><a href="/testdir/testprefer/dir3/">dir3</a></p></li>`,
			`<li><p><a href="/testdir/testprefer/dir4/">Dir-4 title from tag</a></p></li>`,
		},
	}

	notContains := map[string][]string{
		"_index.soyweb": {
			"ignore1",
			"ignore2",
		},
		"testdir/_index.soyweb": {
			"ignore3",
			"ignore4",
		},
		"testdir/testprefer/_index.soyweb": {
			"Dir-3",
			"This should not be copied to target",
		},
	}

	// Sanity checks
	for marker := range contains {
		markerPath := filepath.Join(src, marker)
		assertFs(t, markerPath, false)

		index := formatIndexPath(markerPath)
		_, err := os.Stat(index)
		if err == nil {
			t.Fatalf("unexpected index.html before generator runs")
		}
	}

	err = ssg.Generate(src, dst, title, url, ssg.WithPipelines(soyweb.IndexGenerator))
	if err != nil {
		t.Fatalf("error during ssg generation: %v", err)
	}

	t.Run("should contain expected generated HTML", func(t *testing.T) {
		for marker, entries := range contains {
			markerPath := filepath.Join(dst, marker)
			index := formatIndexPath(markerPath)
			assertFs(t, index, false)

			content, err := os.ReadFile(index)
			if err != nil {
				t.Fatalf("failed to read back index %s: %v", index, err)
			}
			for i := range entries {
				entry := entries[i]
				if strings.Contains(string(content), entry) {
					continue
				}
				t.Log("actual content:\n", string(content))
				t.Fatalf("missing #%d entry '%s' in %s", i+1, entry, index)
			}
		}
	})

	t.Run("ssgignore should work", func(t *testing.T) {
		for marker, entries := range notContains {
			markerPath := filepath.Join(dst, marker)
			index := formatIndexPath(markerPath)
			assertFs(t, index, false)

			content, err := os.ReadFile(index)
			if err != nil {
				t.Fatalf("failed to read back index %s: %v", index, err)
			}
			for i := range entries {
				entry := entries[i]
				if !strings.Contains(string(content), entry) {
					continue
				}
				t.Fatalf("unexpected #%d entry '%s' in %s", i+1, entry, index)
			}
		}
	})

	err = os.RemoveAll(dst)
	if err != nil {
		panic(err)
	}
}

func TestGenerateIndexReverse(t *testing.T) {
	src := "./testdata/myblog/src"
	dst := "./testdata/myblog/dst-test-generate-index-reverse"
	title := "TestTitle"
	url := "https://my.blog.testgenindexreverse"

	err := os.RemoveAll(dst)
	if err != nil {
		panic(err)
	}

	// For reverse index, the order should be reversed compared to default.
	// We'll test that the links appear in reverse alphabetical order.
	// The default order produces entries sorted alphabetically,
	// so reverse should flip them.
	reverseOrdering := map[string]struct {
		first string // Should appear first in the list
		last  string // Should appear last in the list
	}{
		"_index.soyweb": {
			first: `<li><p><a href="/testdir/">testdir Title from tag</a></p></li>`,
			last:  `<li><p><a href="/2022/">twenty and twenty-two (from-tag)</a></p></li>`,
		},
		"2022/_index.soyweb": {
			first: `<li><p><a href="/2022/foo.html">Foo</a></p></li>`,
			last:  `<li><p><a href="/2022/bar/">bar</a></p></li>`,
		},
	}

	err = ssg.Generate(src, dst, title, url, ssg.WithPipelines(soyweb.IndexGeneratorReverse))
	if err != nil {
		t.Fatalf("error during ssg generation with reverse: %v", err)
	}

	t.Run("should reverse order of entries", func(t *testing.T) {
		for marker, ordering := range reverseOrdering {
			markerPath := filepath.Join(dst, marker)
			index := formatIndexPath(markerPath)
			assertFs(t, index, false)

			content, err := os.ReadFile(index)
			if err != nil {
				t.Fatalf("failed to read back index %s: %v", index, err)
			}

			contentStr := string(content)

			// Find positions of first and last entries
			firstPos := strings.Index(contentStr, ordering.first)
			lastPos := strings.Index(contentStr, ordering.last)

			if firstPos == -1 {
				t.Fatalf("expected first entry '%s' not found in %s", ordering.first, index)
			}
			if lastPos == -1 {
				t.Fatalf("expected last entry '%s' not found in %s", ordering.last, index)
			}

			// In reverse, first should come before last
			if firstPos >= lastPos {
				t.Log("actual content:\n", contentStr)
				t.Fatalf("reverse ordering not working in %s: first entry at pos %d, last entry at pos %d", index, firstPos, lastPos)
			}
		}
	})

	err = os.RemoveAll(dst)
	if err != nil {
		panic(err)
	}
}

func TestGenerateIndexModTime(t *testing.T) {
	// Create a temporary directory with controlled modification times
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")
	title := "TestTitle"
	url := "https://my.blog.testgenindexmodtime"

	// Setup test structure with specific modification times
	// We'll create files with known modtimes to test sorting
	now := time.Now()
	oldTime := now.Add(-72 * time.Hour)    // 3 days ago
	mediumTime := now.Add(-24 * time.Hour) // 1 day ago
	recentTime := now.Add(-1 * time.Hour)  // 1 hour ago

	// Create directory structure
	err := os.MkdirAll(filepath.Join(src, "articles"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	// Create articles with different modification times
	// oldest.md - oldest file
	oldestPath := filepath.Join(src, "articles", "oldest.md")
	err = os.WriteFile(oldestPath, []byte("# Oldest Article\n\nContent"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chtimes(oldestPath, oldTime, oldTime)
	if err != nil {
		t.Fatal(err)
	}

	// medium.md - medium age file
	mediumPath := filepath.Join(src, "articles", "medium.md")
	err = os.WriteFile(mediumPath, []byte("# Medium Article\n\nContent"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chtimes(mediumPath, mediumTime, mediumTime)
	if err != nil {
		t.Fatal(err)
	}

	// newest.md - newest file
	newestPath := filepath.Join(src, "articles", "newest.md")
	err = os.WriteFile(newestPath, []byte("# Newest Article\n\nContent"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chtimes(newestPath, recentTime, recentTime)
	if err != nil {
		t.Fatal(err)
	}

	// Create the _index.soyweb marker
	markerPath := filepath.Join(src, "articles", "_index.soyweb")
	err = os.WriteFile(markerPath, []byte("# Articles Index\n\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Generate with ModTime indexer
	err = ssg.Generate(src, dst, title, url, ssg.WithPipelines(soyweb.IndexGeneratorModTime))
	if err != nil {
		t.Fatalf("error during ssg generation with modtime: %v", err)
	}

	t.Run("should sort entries by modification time (oldest first)", func(t *testing.T) {
		indexPath := filepath.Join(dst, "articles", "index.html")
		assertFs(t, indexPath, false)

		content, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("failed to read index: %v", err)
		}

		contentStr := string(content)

		// Expected links in order (oldest to newest)
		oldestLink := `<a href="/articles/oldest.html">Oldest Article</a>`
		mediumLink := `<a href="/articles/medium.html">Medium Article</a>`
		newestLink := `<a href="/articles/newest.html">Newest Article</a>`

		// Find positions
		oldestPos := strings.Index(contentStr, oldestLink)
		mediumPos := strings.Index(contentStr, mediumLink)
		newestPos := strings.Index(contentStr, newestLink)

		// Verify all links exist
		if oldestPos == -1 {
			t.Log("content:\n", contentStr)
			t.Fatal("oldest article link not found")
		}
		if mediumPos == -1 {
			t.Log("content:\n", contentStr)
			t.Fatal("medium article link not found")
		}
		if newestPos == -1 {
			t.Log("content:\n", contentStr)
			t.Fatal("newest article link not found")
		}

		// Verify order: oldest < medium < newest
		if oldestPos >= mediumPos {
			t.Log("content:\n", contentStr)
			t.Fatalf("oldest article (pos %d) should come before medium article (pos %d)", oldestPos, mediumPos)
		}
		if mediumPos >= newestPos {
			t.Log("content:\n", contentStr)
			t.Fatalf("medium article (pos %d) should come before newest article (pos %d)", mediumPos, newestPos)
		}
	})

	t.Run("handles files with same modtime using alphabetical sort", func(t *testing.T) {
		// Create another test directory for same-time files
		sameSrc := t.TempDir()
		sameDst := filepath.Join(t.TempDir(), "dst")

		err := os.MkdirAll(filepath.Join(sameSrc, "posts"), 0o755)
		if err != nil {
			t.Fatal(err)
		}

		sameTime := now.Add(-12 * time.Hour)

		// Create files with identical modtimes
		for _, name := range []string{"zebra.md", "alpha.md", "beta.md"} {
			path := filepath.Join(sameSrc, "posts", name)
			title := strings.TrimSuffix(name, ".md")
			err = os.WriteFile(path, fmt.Appendf(nil, "# %s\n\nContent", title), 0o644)
			if err != nil {
				t.Fatal(err)
			}
			err = os.Chtimes(path, sameTime, sameTime)
			if err != nil {
				t.Fatal(err)
			}
		}

		// Create marker
		markerPath := filepath.Join(sameSrc, "posts", "_index.soyweb")
		err = os.WriteFile(markerPath, []byte("# Posts\n\n"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		err = ssg.Generate(sameSrc, sameDst, title, url, ssg.WithPipelines(soyweb.IndexGeneratorModTime))
		if err != nil {
			t.Fatalf("error during generation: %v", err)
		}

		indexPath := filepath.Join(sameDst, "posts", "index.html")
		content, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("failed to read index: %v", err)
		}

		contentStr := string(content)

		// When modtimes are equal, should fall back to alphabetical order
		alphaPos := strings.Index(contentStr, `<a href="/posts/alpha.html">alpha</a>`)
		betaPos := strings.Index(contentStr, `<a href="/posts/beta.html">beta</a>`)
		zebraPos := strings.Index(contentStr, `<a href="/posts/zebra.html">zebra</a>`)

		if alphaPos == -1 || betaPos == -1 || zebraPos == -1 {
			t.Log("content:\n", contentStr)
			t.Fatal("not all links found")
		}

		if alphaPos >= betaPos || betaPos >= zebraPos {
			t.Log("content:\n", contentStr)
			t.Fatalf("files with same modtime should be alphabetically sorted: alpha(%d) < beta(%d) < zebra(%d)",
				alphaPos, betaPos, zebraPos)
		}
	})
}

func formatIndexPath(marker string) string {
	marker = filepath.Dir(marker)
	return filepath.Join(marker, "index.html")
}
