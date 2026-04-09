package utils

import (
	"reflect"
	"testing"
)

func TestSplitHTMLContentChunks(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []ContentChunk
	}{
		{
			name:    "plain text paragraph",
			content: "<p>你好，微信用户</p>",
			want: []ContentChunk{
				{Type: ContentChunkTypeText, Content: "你好，微信用户"},
			},
		},
		{
			name:    "single image",
			content: `<p><img src="https://example.com/a.png" alt="a"></p>`,
			want: []ContentChunk{
				{Type: ContentChunkTypeImage, Content: "https://example.com/a.png"},
			},
		},
		{
			name:    "mixed text and images keep order",
			content: `<p>第一段</p><p><img src="https://example.com/1.png"></p><p>第二段<img src="https://example.com/2.png">第三段</p>`,
			want: []ContentChunk{
				{Type: ContentChunkTypeText, Content: "第一段"},
				{Type: ContentChunkTypeImage, Content: "https://example.com/1.png"},
				{Type: ContentChunkTypeText, Content: "第二段"},
				{Type: ContentChunkTypeImage, Content: "https://example.com/2.png"},
				{Type: ContentChunkTypeText, Content: "第三段"},
			},
		},
		{
			name:    "normalize blank lines and spaces",
			content: "<div>  第一行  <br><br>   第二行 </div>",
			want: []ContentChunk{
				{Type: ContentChunkTypeText, Content: "第一行\n\n第二行"},
			},
		},
		{
			name:    "ignore image without src",
			content: `<p>前文</p><img alt="missing-src"><p>后文</p>`,
			want: []ContentChunk{
				{Type: ContentChunkTypeText, Content: "前文"},
				{Type: ContentChunkTypeText, Content: "后文"},
			},
		},
		{
			name:    "empty html returns empty chunks",
			content: "<p><br></p>",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitHTMLContentChunks(tt.content)
			if err != nil {
				t.Fatalf("SplitHTMLContentChunks() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("SplitHTMLContentChunks() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
