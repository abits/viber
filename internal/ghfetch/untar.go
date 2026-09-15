package ghfetch

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"testing/fstest"
)

// Options configures the safety limits UntarGz enforces while extracting.
// The zero value enforces none: no size limit, no path validation. That
// matches how internal/updater extracted release archives before this
// package existed, which is safe there because it never writes an entry's
// tar path to disk — only its content, to a destination the caller already
// fixed.
type Options struct {
	// MaxUncompressedBytes caps the total size of extracted regular-file
	// content. Zero means unlimited.
	MaxUncompressedBytes int64
	// RejectUnsafePaths rejects any entry whose name is not fs.ValidPath or
	// contains a backslash. fs.ValidPath alone is not enough: it treats
	// backslash as an ordinary character rather than a separator (see its
	// doc comment), but callers that later join such a name with
	// path/filepath — as templates.Render does — get OS-specific separator
	// handling, and on Windows that means a name like "..\evil.txt" walks
	// out of the intended destination. Set this for any archive whose
	// entries will be written to disk under their own paths.
	RejectUnsafePaths bool
}

// UntarGz decompresses and unpacks a gzip-compressed tar stream into an
// in-memory fs.FS containing every regular-file entry, keyed by its full tar
// path. Non-regular entries (directories, symlinks, ...) are skipped.
func UntarGz(r io.Reader, opts Options) (fs.FS, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	m := fstest.MapFS{}
	var total int64
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if opts.RejectUnsafePaths && (!fs.ValidPath(hdr.Name) || strings.ContainsRune(hdr.Name, '\\')) {
			return nil, fmt.Errorf("archive contains an unsafe path %q", hdr.Name)
		}
		buf, newTotal, err := readEntry(tr, opts.MaxUncompressedBytes, total)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", hdr.Name, err)
		}
		total = newTotal
		// raw tar mode is truncated to permission bits (0o777) so any overflow is intentional.
		m[hdr.Name] = &fstest.MapFile{Data: buf, Mode: fs.FileMode(hdr.Mode) & 0o777}
	}
	return m, nil
}

// readEntry reads the current tar entry's content, enforcing max (the
// running total across the whole archive so far) when max is positive. It
// returns the new running total so the caller can thread it into the next
// call. The limit is enforced against bytes actually read, not the entry's
// declared Size, since a tar header's Size is attacker-controlled and
// cannot be trusted.
func readEntry(tr *tar.Reader, max, total int64) (buf []byte, newTotal int64, err error) {
	if max <= 0 {
		buf, err = io.ReadAll(tr)
		return buf, total + int64(len(buf)), err
	}
	remaining := max - total
	buf, err = io.ReadAll(io.LimitReader(tr, remaining+1))
	if err != nil {
		return nil, total, err
	}
	newTotal = total + int64(len(buf))
	if newTotal > max {
		return nil, newTotal, fmt.Errorf("archive exceeds maximum uncompressed size of %d bytes", max)
	}
	return buf, newTotal, nil
}
