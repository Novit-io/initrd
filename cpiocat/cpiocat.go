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

		mode := hdr.FileInfo().Mode()

		if mode&os.ModeSymlink != 0 {
			// symlink target must be written after
			hdr.Size = int64(len(hdr.Linkname))
		}

		err = cout.WriteHeader(hdr)
		if err != nil {
			return
		}

		if mode.IsRegular() {
			_, err = io.Copy(cout, cin)

		} else if mode&os.ModeSymlink != 0 {
			_, err = cout.Write([]byte(hdr.Linkname))
		}

		if err != nil {
			return
		}
	}

	for _, file := range filesToAppend {
		err = func() (err error) {
			stat, err := os.Lstat(file)
			if err != nil {
				return
			}

			link := ""
			if stat.Mode()&os.ModeSymlink != 0 {
				link, err = os.Readlink(file)
				if err != nil {
					return
				}
			}

			hdr, err := cpio.FileInfoHeader(stat, link)
			if err != nil {
				return
			}

			hdr.Name = file

			cout.WriteHeader(hdr)

			if stat.Mode().IsRegular() {
				var f *os.File
				f, err = os.Open(file)
				if err != nil {
					return
				}
				defer f.Close()

				_, err = io.Copy(cout, f)

			} else if stat.Mode()&os.ModeSymlink != 0 {
				_, err = cout.Write([]byte(link))
			}
			return
		}()
		if err != nil {
			return
		}
	}

	err = cout.Close()
	return
}
