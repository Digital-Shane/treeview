package s3

import (
	"path/filepath"
	"testing"

	"github.com/Digital-Shane/treeview/extensions/s3/internal/localstack"
)

const (
	_myTestBucket = "example.com"
	_cS3          = _cS3URI + _myTestBucket
	_cs3Testdata  = _cS3 + "/testdata"
	_c100K        = "sample100K.golden"
	_c1M          = "sample1M.golden"
	_c1Mg         = "sample1M.golden1"
	_cGolden100K  = _cs3Testdata + "/golden/" + _c100K
	_cGolden1M    = _cs3Testdata + "/golden/" + _c1Mg
	K             = 1024
)

var _goldenDirPath string

func TestMain(m *testing.M) {
	ok := localstack.IsRunning()
	if !ok {
		panic("localstack not running")
	}
	isPanic(localstack.Use())
	defer func() { _ = localstack.UseNot() }()

	_goldenDirPath = filepath.Join("..", "testfixtures")
	_ = localstack.CreateBucket(_myTestBucket, localstack.WithNoErrorIfExist())
	// sample100K.golden (resp. sample1M.golden) is a 100K (resp. 1M) random byte file used for test purpose.
	_ = localstack.PutObject(_cGolden100K, filepath.Join(_goldenDirPath, _c100K))
	_ = localstack.PutObject(_cGolden1M, filepath.Join(_goldenDirPath, _c1M))
	_ = localstack.PutObject(_cs3Testdata+"/golden/recurse/"+_c100K, filepath.Join(_goldenDirPath, _c100K))
	m.Run()
}

func isPanic(err error) {
	if err != nil {
		panic(err)
	}
}
