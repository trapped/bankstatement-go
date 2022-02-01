package bankstatement

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
)

// Decoder is the interface for intermediate or container file format readers,
// for example for reading compressed or wrapped files.
type Decoder interface {
	// Wrap chains the current decoder on top of a ReaderAt, yielding a new, wrapped ReaderAt.
	Wrap(Reader) (Reader, error)
}

// MIMEMultipartDecoder scans and reads files wrapped with a single MIME/Multipart UUID boundary.
// The boundary is automatically detected.
type MIMEMultipartDecoder struct {
}

// findBoundary scans the reader looking for a boundary start marker.
func (d *MIMEMultipartDecoder) findBoundary(r Reader) (boundary string, err error) {
	// reset reader when done
	defer r.Seek(0, io.SeekStart)
	s := bufio.NewScanner(r)
	for s.Scan() && boundary == "" {
		line := s.Bytes()
		if bytes.HasPrefix(line, []byte("--uuid:")) {
			boundary = string(bytes.TrimPrefix(bytes.TrimSpace(line), []byte("--")))
			return
		}
	}
	err = s.Err()
	if boundary == "" && err != nil {
		err = errors.New("boundary not found")
	}
	return
}

// Wrap wraps r with a MIME/Multipart reader.
// The data from the multipart section is read and buffered.
func (d *MIMEMultipartDecoder) Wrap(r Reader) (ReaderSize, error) {
	boundary, err := d.findBoundary(r)
	if err != nil {
		return nil, err
	}
	mr := multipart.NewReader(r, boundary)
	part, err := mr.NextPart()
	if err != nil {
		return nil, err
	}
	partdata, err := ioutil.ReadAll(part)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(partdata), nil
}
