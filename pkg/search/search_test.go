package search

import (
	"testing"

	"github.com/fu9zhou/finish-bit/pkg/operation"
)

func TestSearchRanksAliasAheadOfDescription(t *testing.T) {
	index := New([]operation.Definition{
		{ID: "video.trim", Summary: "Trim video", Description: "Change a media file", Aliases: []string{"视频裁剪"}},
		{ID: "video.compress", Summary: "Compress video", Description: "裁剪后也可以压缩视频"},
	})
	matches := index.Search("视频裁剪", 5)
	if len(matches) == 0 {
		t.Fatal("search returned no matches")
	}
	if matches[0].ID != "video.trim" {
		t.Fatalf("top result = %q, want video.trim", matches[0].ID)
	}
}

func TestSearchHonorsLimit(t *testing.T) {
	index := New([]operation.Definition{{ID: "text.sort", Summary: "Sort text"}, {ID: "text.count", Summary: "Count text"}})
	if matches := index.Search("text", 1); len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
}

func TestSearchDropsWeakCrossDomainMatches(t *testing.T) {
	index := New([]operation.Definition{
		{ID: "video.compress", Summary: "Compress video", Aliases: []string{"压缩视频"}, Tags: []string{"video"}},
		{ID: "video.trim", Summary: "Trim video", Aliases: []string{"视频裁剪"}, Tags: []string{"video"}},
		{ID: "json.minify", Summary: "Minify JSON", Aliases: []string{"压缩 json"}, Tags: []string{"json"}},
	})
	matches := index.Search("裁剪并压缩视频", 5)
	for _, match := range matches {
		if match.ID == "json.minify" {
			t.Fatal("weak cross-domain match was retained")
		}
	}
}
