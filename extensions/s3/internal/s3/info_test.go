package s3

import (
	"context"
	"io/fs"
	"slices"
	"testing"
)

func Test_ReadDir(t *testing.T) {
	tests := []struct {
		path       string
		expSuccess bool
		minLength  int
		contain    []string
	}{
		{_cs3Testdata + "/golden", true, 2, []string{_c100K, _c1Mg}},
		{_cs3Testdata, true, 1, []string{"golden"}},
		{"badBucket/golden", false, 0, nil},
	}
	for _, tt := range tests {
		a, err := ReadDir(context.Background(), tt.path)
		if tt.expSuccess != (err == nil) {
			t.Fatal(err)
		}
		if err == nil {
			if len(a) < tt.minLength {
				t.Fatalf("expected minimal length %d, got %d", tt.minLength, len(a))
			}
			var as []string
			for _, dir := range a {
				as = append(as, dir.Name())
			}
			for _, name := range tt.contain {
				if !slices.Contains(as, name) {
					t.Errorf("expected %s to contain %s", as, name)
				}
			}
		}
	}
	a, err := ReadDir(context.Background(), Join(_cs3Testdata, "golden", "recurse"))
	isPanic(err)
	if a[0].IsDir() != false {
		t.Error("expected dir to be false")
	}
	if a[0].Name() != _c100K {
		t.Errorf("expected %s to be %s", a[0].Name(), _c100K)
	}

	s, err := a[0].Info()
	isPanic(err)
	if s.Size() != int64(100*K) {
		t.Errorf("expected %d to be %d", s.Size(), 100*K)
	}
}

func Test_DirEntry_is_interface(t *testing.T) {
	var _ fs.DirEntry = (*DirEntry)(nil)
}

func Test_IsDir(t *testing.T) {
	tests := []struct {
		path       string
		expSuccess bool
	}{
		{_cs3Testdata + "/golden", true},
		{_cs3Testdata + "/golden/", true},
		{_cs3Testdata + "/golden/sample", false},
		{_cs3Testdata, true},
		{"s3://bad", false},
	}
	for i, tt := range tests {
		if tt.expSuccess != IsDir(context.Background(), tt.path) {
			t.Errorf("%d: expected %v, got %v", i, tt.expSuccess, IsDir(context.Background(), tt.path))
		}
	}
}

func Test_Base(t *testing.T) {
	tests := []struct {
		path   string
		result string
	}{
		{"s3://bucket/prefix/object/", "object"},
		{"s3://bucket/prefix", "prefix"},
		{"s3://bucket", ""},
		{"", "."},
	}
	for _, tt := range tests {
		if tt.result != Base(tt.path) {
			t.Errorf("%s: expected %s, got %s", tt.path, tt.result, tt.result)
		}
	}
}
