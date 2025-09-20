package localstack

import (
	"math/rand/v2"
	"path/filepath"
	"testing"
)

var (
	// `sample100K.golden` is a 100K-byte file with random bytes.
	src = filepath.Join("..", "testfixtures", "sample100K.golden")
)

func TestMain(m *testing.M) {
	isPanic(Use())
	defer func() { _ = UseNot() }()
	_ = m.Run()
}

func TestIsRunning(t *testing.T) {
	if IsRunning() == false {
		t.Fatal("expected true")
	}
}

func TestCreateBucket(t *testing.T) {
	bkt := "tst" + randomSmallCapsID()
	if err := CreateBucket(bkt); err != nil {
		t.Fatal(err)
	}
	if err := CreateBucket(bkt); err == nil {
		t.Error(err)
	}
	if err := CreateBucket(bkt, WithNoErrorIfExist()); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBucket(bkt); err != nil {
		t.Error(err)
	}
	if err := DeleteBucket("s3://bad"); err == nil {
		t.Error("should generate an error")
	}
	if err := CreateBucket("tstA"); err == nil {
		t.Error("should generate an error")
	}
}

func TestCreateBucketAt(t *testing.T) {
	bkt := "tst" + randomSmallCapsID()
	const cRegion = "us-west-2"
	if err := CreateBucketAt(bkt, cRegion); err != nil {
		t.Fatal(err)
	}
	if err := DeleteBucketAt(bkt, cRegion); err != nil {
		t.Error(err)
	}
}

func TestPutObject(t *testing.T) {
	bkt := "tst" + randomSmallCapsID()
	isPanic(CreateBucket(bkt))
	defer func() { _ = DeleteBucket(bkt) }()

	if err := PutObject("s3://"+bkt+"/test", src); err != nil {
		t.Fatal(err)
	}
	if err := PutObject("s3://"+bkt+"/test", src); err != nil {
		t.Error(err)
	}
}

func TestDeleteBucket(t *testing.T) {
	bkt := "tst" + randomSmallCapsID()
	isPanic(CreateBucket(bkt))
	isPanic(PutObject("s3://"+bkt+"/test", src))

	if err := DeleteBucket(bkt); err != nil {
		t.Error(err)
	}
}

func TestUseNot(t *testing.T) {
	defer func() { _ = Use() }()

	if InUse() == false {
		t.Fatal("InUse() should return true")
	}
	if err := UseNot(); err != nil {
		t.Fatalf("UseNot() should not return error, but got %v", err)
	}
	if InUse() == true {
		t.Error("UseNot() should return false")
	}
	if err := Use(); err != nil {
		t.Fatalf("Use should not return error, but got %v", err)
	}
	if InUse() == false {
		t.Fatal("InUse() should return true")
	}
}

func isPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func randomSmallCapsID() string {
	const sizeID = 16
	size := sizeID
	var buffer []byte
	choice := []byte("abcdefghijklmnopqrstuvwxyz01234567890")
	choiceSize := len(choice)
	for i := 0; i < size; i++ {
		// generates the characters
		s := rand.IntN(choiceSize)
		buffer = append(buffer, choice[s])
	}
	return string(buffer)
}
