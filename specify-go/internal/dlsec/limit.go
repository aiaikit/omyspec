package dlsec

import (
	"fmt"
	"io"
)

// Size caps (mirrors Python's _download_security constants).
const (
	MaxDownloadBytes    = 50 * 1024 * 1024 // 50 MiB
	MaxJSONCatalogBytes = 5 * 1024 * 1024  // 5 MiB
)

// ReadLimited reads up to n+1 bytes; returns (data, truncated, err).
// If truncated is true, the stream was longer than n bytes and the
// returned data is the prefix of length n.
func ReadLimited(r io.Reader, n int64) ([]byte, bool, error) {
	if n < 0 {
		return nil, false, fmt.Errorf("dlsec: ReadLimited cap %d < 0", n)
	}
	lr := io.LimitReader(r, n+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > n {
		return data[:n], true, nil
	}
	return data, false, nil
}
