package s3

import (
	"context"
	"io/fs"
	"testing"
)

func TestInfo1(t *testing.T) {
	oi, err := Info(context.Background(), _cs3Testdata)
	if err != nil {
		t.Fatal(err)
	}
	if oi == nil {
		t.Fatal("Info() returned nil")
	}
	if oi.IsDir() == false {
		t.Error("IsDir() returned false")
	}
	if oi.Mode() != fs.ModeDir|defaultMode {
		t.Errorf("expected %v got %v", fs.ModeDir|defaultMode, oi.Mode())
	}
	if oi.Bucket() != _myTestBucket {
		t.Errorf("expected %v got %v", _myTestBucket, oi.Bucket())
	}

	oi, err = Info(context.Background(), _cS3)
	if err != nil {
		t.Fatal(err)
	}
	if oi == nil {
		t.Fatal("Info() returned nil")
	}
	if oi.IsDir() == false {
		t.Error("IsDir() returned false")
	}

	oi, err = Info(context.Background(), _cGolden100K)
	if err != nil {
		t.Fatal(err)
	}
	if oi == nil {
		t.Fatal("Info() returned nil")
	}
	if oi.IsDir() == true {
		t.Error("IsDir() returned true")
	}
	if oi.Name() != _c100K {
		t.Error("expected " + _c100K + ", got " + oi.Name())
	}
	if int64(100*K) != oi.Size() {
		t.Errorf("expected %d got %d", int64(100*K), oi.Size())
	}

	if oi.Mode() != fs.ModeIrregular {
		t.Errorf("expected %v got %v", fs.ModeIrregular, oi.Mode())
	}
}
