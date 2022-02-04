package cpiocat

import (
	"io"
	"os"

	"github.com/cavaliergopher/cpio"
)

func Append(out io.Writer, in io.Reader, filesToAppend []string) (err error) {
	cout := cpio.NewWriter(out)

	cin := cpio.NewReader(in)

	for {
		var hdr *cpio.Header
		hdr, err = cin.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		err = cout.WriteHeader(hdr)
		if err != nil {
			return
		}

		_, err = io.Copy(cout, cin)
		if err != nil {
			return
		}
	}

	for _, file := range filesToAppend {
		err = func() (err error) {
			stat, err := os.Stat(file)
			if err != nil {
				return
			}

			hdr := &cpio.Header{
				Name: file,
				Size: stat.Size(),
				Mode: cpio.FileMode(stat.Mode()),
			}

			cout.WriteHeader(hdr)

			f, err := os.Open(file)
			if err != nil {
				return
			}
			defer f.Close()

			_, err = io.Copy(cout, f)
			return
		}()
		if err != nil {
			return
		}
	}

	err = cout.Close()
	return
}
