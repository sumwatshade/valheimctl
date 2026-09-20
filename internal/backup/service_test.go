package backup

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

type testFS struct {
	files map[string][]byte
	dirs  map[string]struct{}
}

func newTestFS(m fstest.MapFS) *testFS {
	fsys := &testFS{files: map[string][]byte{}, dirs: map[string]struct{}{}}
	for path, file := range m {
		clean := filepath.ToSlash(filepath.Clean(path))
		if file == nil {
			fsys.dirs[clean] = struct{}{}
			continue
		}
		fsys.files[clean] = append([]byte(nil), file.Data...)
		for dir := filepath.Dir(clean); dir != "." && dir != "/"; dir = filepath.Dir(dir) {
			fsys.dirs[filepath.ToSlash(filepath.Clean(dir))] = struct{}{}
		}
	}
	return fsys
}

func (f *testFS) ReadDir(name string) ([]fs.DirEntry, error) {
	prefix := filepath.ToSlash(filepath.Clean(name))
	seen := map[string]struct{}{}
	for path := range f.files {
		if dir := filepath.Dir(path); dir == prefix {
			seen[filepath.Base(path)] = struct{}{}
		}
	}
	for dir := range f.dirs {
		if dir == prefix {
			continue
		}
		if strings.HasPrefix(dir, prefix+"/") {
			seen[filepath.Base(dir)] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return nil, fs.ErrNotExist
	}
	out := make([]fs.DirEntry, 0, len(seen))
	for item := range seen {
		full := filepath.ToSlash(filepath.Join(prefix, item))
		out = append(out, testDirEntry{name: item, isDir: f.isDir(full)})
	}
	return out, nil
}

func (f *testFS) isDir(path string) bool {
	_, ok := f.dirs[path]
	return ok
}

func (f *testFS) Open(name string) (fs.File, error) {
	p := filepath.ToSlash(filepath.Clean(name))
	if data, ok := f.files[p]; ok {
		return &testFile{data: append([]byte(nil), data...), pos: 0}, nil
	}
	if _, ok := f.dirs[p]; ok {
		return &testDirFile{name: p, entries: f.listDir(p)}, nil
	}
	return nil, fs.ErrNotExist
}

func (f *testFS) ReadFile(name string) ([]byte, error) {
	p := filepath.ToSlash(filepath.Clean(name))
	if data, ok := f.files[p]; ok {
		return append([]byte(nil), data...), nil
	}
	return nil, fs.ErrNotExist
}

func (f *testFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	p := filepath.ToSlash(filepath.Clean(name))
	f.files[p] = append([]byte(nil), data...)
	for dir := filepath.Dir(p); dir != "." && dir != "/"; dir = filepath.Dir(dir) {
		f.dirs[filepath.ToSlash(filepath.Clean(dir))] = struct{}{}
	}
	return nil
}

func (f *testFS) MkdirAll(path string, perm fs.FileMode) error {
	p := filepath.ToSlash(filepath.Clean(path))
	for dir := p; dir != "." && dir != "/"; dir = filepath.Dir(dir) {
		f.dirs[filepath.ToSlash(filepath.Clean(dir))] = struct{}{}
	}
	return nil
}

func (f *testFS) RemoveAll(path string) error {
	p := filepath.ToSlash(filepath.Clean(path))
	for file := range f.files {
		if file == p || strings.HasPrefix(file, p+"/") {
			delete(f.files, file)
		}
	}
	for dir := range f.dirs {
		if dir == p || strings.HasPrefix(dir, p+"/") {
			delete(f.dirs, dir)
		}
	}
	return nil
}

func (f *testFS) Stat(name string) (fs.FileInfo, error) {
	p := filepath.ToSlash(filepath.Clean(name))
	if data, ok := f.files[p]; ok {
		return testFileInfo{name: filepath.Base(p), size: int64(len(data)), mode: 0}, nil
	}
	if _, ok := f.dirs[p]; ok {
		return testFileInfo{name: filepath.Base(p), mode: fs.ModeDir}, nil
	}
	return nil, fs.ErrNotExist
}

func (f *testFS) listDir(name string) []fs.DirEntry {
	prefix := filepath.ToSlash(filepath.Clean(name))
	seen := map[string]struct{}{}
	for path := range f.files {
		if dir := filepath.Dir(path); dir == prefix {
			seen[filepath.Base(path)] = struct{}{}
		}
	}
	for dir := range f.dirs {
		if dir == prefix {
			continue
		}
		if strings.HasPrefix(dir, prefix+"/") {
			seen[filepath.Base(dir)] = struct{}{}
		}
	}
	out := make([]fs.DirEntry, 0, len(seen))
	for item := range seen {
		full := filepath.ToSlash(filepath.Join(prefix, item))
		out = append(out, testDirEntry{name: item, isDir: f.isDir(full)})
	}
	return out
}

type testDirEntry struct {
	name  string
	isDir bool
}

func (d testDirEntry) Name() string { return d.name }
func (d testDirEntry) IsDir() bool  { return d.isDir }
func (d testDirEntry) Type() fs.FileMode {
	if d.isDir {
		return fs.ModeDir
	}
	return 0
}
func (d testDirEntry) Info() (fs.FileInfo, error) {
	return testFileInfo{name: d.name, mode: map[bool]fs.FileMode{true: fs.ModeDir, false: 0}[d.isDir]}, nil
}

type testFile struct {
	data []byte
	pos  int
}

func (f *testFile) Stat() (fs.FileInfo, error) {
	return testFileInfo{name: "memory", size: int64(len(f.data))}, nil
}
func (f *testFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.pos:])
	f.pos += n
	if f.pos >= len(f.data) {
		return n, io.EOF
	}
	return n, nil
}
func (f *testFile) Close() error { return nil }

type testDirFile struct {
	name    string
	entries []fs.DirEntry
	pos     int
}

func (d *testDirFile) Stat() (fs.FileInfo, error) {
	return testFileInfo{name: filepath.Base(d.name), mode: fs.ModeDir}, nil
}
func (d *testDirFile) Read([]byte) (int, error) { return 0, fs.ErrInvalid }
func (d *testDirFile) Close() error             { return nil }
func (d *testDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.pos >= len(d.entries) {
		if n <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}
	end := d.pos + n
	if n <= 0 || end > len(d.entries) {
		end = len(d.entries)
	}
	out := d.entries[d.pos:end]
	d.pos = end
	return out, nil
}

type testFileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

func (i testFileInfo) Name() string       { return i.name }
func (i testFileInfo) Size() int64        { return i.size }
func (i testFileInfo) Mode() fs.FileMode  { return i.mode }
func (i testFileInfo) ModTime() time.Time { return time.Time{} }
func (i testFileInfo) IsDir() bool        { return i.mode.IsDir() }
func (i testFileInfo) Sys() any           { return nil }

func TestBackupCreateListAndApply(t *testing.T) {
	root := t.TempDir()
	svc := NewService(root)
	worldDir := filepath.Join(root, "worlds_local", "Dedicated")
	if err := os.MkdirAll(filepath.Join(worldDir, "chunks"), 0o755); err != nil {
		t.Fatalf("mkdir world dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worldDir, "level.dat"), []byte("live"), 0o600); err != nil {
		t.Fatalf("write level.dat: %v", err)
	}
	if err := os.WriteFile(filepath.Join(worldDir, "chunks", "a.chunk"), []byte("chunk-data"), 0o600); err != nil {
		t.Fatalf("write chunk: %v", err)
	}

	if err := svc.Create("alpha"); err != nil {
		t.Fatalf("backup create returned error: %v", err)
	}

	items, err := svc.List()
	if err != nil {
		t.Fatalf("backup list returned error: %v", err)
	}
	if len(items) == 0 || items[0].Name != "alpha" {
		t.Fatalf("backup list = %#v; want alpha at the front", items)
	}

	if err := os.WriteFile(filepath.Join(worldDir, "level.dat"), []byte("changed"), 0o600); err != nil {
		t.Fatalf("overwrite level.dat: %v", err)
	}
	if err := svc.Apply("alpha"); err != nil {
		t.Fatalf("backup apply returned error: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(worldDir, "level.dat"))
	if err != nil {
		t.Fatalf("read restored level.dat: %v", err)
	}
	if string(b) != "live" {
		t.Fatalf("restored world file = %q; want %q", string(b), "live")
	}
}

func TestBackupApplyRejectsMissingBackup(t *testing.T) {
	root := t.TempDir()
	svc := NewService(root)
	if err := svc.Apply("missing"); err == nil {
		t.Fatal("Apply() succeeded for missing backup; want error")
	}
}

func TestBackupServiceAcceptsInjectedFilesystem(t *testing.T) {
	root := "/valheim"
	fsys := newTestFS(fstest.MapFS{
		"/valheim/worlds_local/Dedicated/level.dat":                   &fstest.MapFile{Data: []byte("live")},
		"/valheim/worlds_local/Dedicated/chunks/a.chunk":              &fstest.MapFile{Data: []byte("chunk-a")},
		"/valheim/.valheimctl/backups/alpha/Dedicated/level.dat":      &fstest.MapFile{Data: []byte("backup-live")},
		"/valheim/.valheimctl/backups/alpha/Dedicated/chunks/a.chunk": &fstest.MapFile{Data: []byte("backup-chunk")},
		"/valheim/.valheimctl/backups/alpha/meta.json":                &fstest.MapFile{Data: []byte(`{"name":"alpha","created_at":"2024-01-01T00:00:00Z"}`)},
	})

	svc := NewService(root, fsys)
	if err := svc.Apply("alpha"); err != nil {
		t.Fatalf("apply with injected fs returned error: %v", err)
	}
	if got, err := fsys.ReadFile("/valheim/worlds_local/Dedicated/level.dat"); err != nil || string(got) != "backup-live" {
		t.Fatalf("restored level.dat = %q, err = %v; want backup-live", string(got), err)
	}
}
