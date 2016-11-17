package plotHttp

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/index.html": {
		local:   "static/index.html",
		size:    942,
		modtime: 1479426068,
		compressed: `
H4sIAAAJbogA/6yTse7bIBDGdz8F9f7nkiZTRbwkXbo0Urt0qghcbFwMLneOmrcvjp2oiqI2qjIBH/D7
juNOvdt93n79tv8oGu58VajrgNrmgR17rHaatdj7yKRgUopCkUmuZ0HJbMqGuacPAMaGlqTxcbBHrxNK
EzvQrf4F3h0Ikg42dib6mGAh13I9K9tRkS2VlYIJWz3mjyhZx1h71L2jO3z7c8B0hpVcyuW8kJ0LrwIP
DjL4/Y39Nrhn8XaV0xJTnSfytH721j+y2SEnZ+itTrpv8gRyaIsc3axf5QduhfIu/BAJ/aYkPnukBpFL
0SQ8/m9OuMEOCaiLkZuARH9kyVC2f870hW8eXQXclyqxZmegz9V8/wd/i2++dtmY31MomLvkEO15XFt3
EsZrok1pc8t8NzGwdgHTaJM3xzMhzn6f9El/mQJzJKwjffBopYLbidFhQmenS1v+DgAA//9xVcWMrgMA
AA==
`,
	},

	"/plot.js": {
		local:   "static/plot.js",
		size:    1788,
		modtime: 1479426248,
		compressed: `
H4sIAAAJbogA/3xV30/kNhB+Tv6KaYoURyxh92ilKiwnneB0nNRTq5aqD6dT5EtM4uK1V46zB6X7v3fG
TtiEQh/Ixt9888PfTIb4159/uSn//Hh1cw0X8NNyGYDr9x8/XN8gckbIuz9+f3+Fh1uuOhHHt72unDQa
+m3NnSi5UuVGOCurrpS6ljtZ94g9sAwe46g+y//qjGbJ6cBJFjBGYAOUIS86ygWv2hGakOQCNN8IHy2K
hqRGiyEp80Y07fGBf/v/VvicPK8LKXxaFJ1DssE9uJaEe/8FeMpr2SZ6PJfgeSqkfhokgIkGB3jCnWSe
KfFqdf8jyCs+FHerjLvCA/b785d4rIjM85bsuPJtG/mf5Rd0eQScAHFfQGD0oqAf2PtCotNTSCl/Wmjx
DdBLsPTNcvXDyWp1slylWRxHjttGOIyEzzIcShypscmfPuRUStlYvm2xnVSBk05hHn8PPJK9gLEsQr7J
2rUFHEadwFbIpnUDGuad4JCygPBLSGWUsQVYrmuzuaQDe1T9RmrTSfdQ0I3sXbrPiHtf8qoSXUcOqVci
JfhhCntZPKykvhM15rI9Vf6sVR22SDfXvGtZmI4dt9DiEcVZLkjgqrULUEKfx5G8Bdbl+N44tF8gIwMr
XG+1d0HGrbHAZPBFHr6M/HOQsPZxQB4fh+HC0ACeU7XcXppavHNMZuckHJWAJsb823oNP2Zw4tNkcExF
PbH+wWzngF2/NHonrANn4OzNV+lwSpxohMVLx9GszqkCL40AVTeIMBFomA5ZI5yUfq7LMsFqiBpHnVBk
+J4QWQe1jhii2UEyUszfHO+K5HSNKZF8kaTeCR9pApXiXXeR+Pg0Ycnb9Sny3qb07bI0zGZltONSC5tm
Od9uha4ZBUXxPIvS4ktuRSf/5l+V8EMc4W41vxGEo5zkhww0KFFjJc7Jarn0JzvQnj5HsRPaLaCXw2aI
5ku8lzk55GHkD4Rx7Y92/52Qeb8YtkeUV0pWd8yZplGi3PK+o3tMuoa3mfVsygzb7+nfx3fhbUZXxmzL
2VIadia1aHDIXlys1FV3IzfC9I69EGZBci39J3X0kj2L/w0AAP//JyelsPwGAAA=
`,
	},

	"/style.css": {
		local:   "static/style.css",
		size:    282,
		modtime: 1479425732,
		compressed: `
H4sIAAAJbogA/2yPQWvEIBCFz+uvEPbSQl3Ssnsxd/9GGeMkkdiM6ISElP73Kml7aRje5X0+fE/cHDC8
dzQz+BmT/BSXPhCwlgF7bsWX+HkSA/E/Ki4WumlItMxOy2v/qNcWl5LDpOVr3GSm4J28GmP+gErg/JK1
fMStmOqDdnVOVrST53NoaVN5BEfr8dFb0b0oDRaempd6t/vzMWHxKmH2O9iAasQQj6m/PWvUETOWok1j
auY7AAD//zltTJkaAQAA
`,
	},

	"/": {
		isDir: true,
		local: "static",
	},
}
