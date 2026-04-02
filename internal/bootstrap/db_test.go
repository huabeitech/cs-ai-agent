package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteFilePath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		dsn  string
		want string
	}{
		{
			name: "plain relative path",
			dsn:  "./data/app.db",
			want: "./data/app.db",
		},
		{
			name: "file uri with query",
			dsn:  "file:./data/app.db?_busy_timeout=5000",
			want: "./data/app.db",
		},
		{
			name: "memory dsn",
			dsn:  "file::memory:?cache=shared",
			want: "",
		},
		{
			name: "memory alias",
			dsn:  ":memory:",
			want: "",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := sqliteFilePath(tt.dsn); got != tt.want {
				t.Fatalf("sqliteFilePath(%q) = %q, want %q", tt.dsn, got, tt.want)
			}
		})
	}
}

func TestEnsureSQLiteDir(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	dbPath := filepath.Join(baseDir, "nested", "app.db")
	dsn := "file:" + dbPath + "?_busy_timeout=5000"

	if err := ensureSQLiteDir(dsn); err != nil {
		t.Fatalf("ensureSQLiteDir() error = %v", err)
	}

	if info, err := os.Stat(filepath.Dir(dbPath)); err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	} else if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", filepath.Dir(dbPath))
	}
}
