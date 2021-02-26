package migrations

import (
	"context"
	"io"
	"os"
	"strings"
)

const (
	// Current dirstibution to fetch migrations from
	CurrentIpfsDist = "/ipfs/Qme8pJhBidEUXRdpcWLGR2fkG5kdwVnaMh3kabjfP8zz7Y"
	// Distribution IPNS path.  Default for fetchers.
	IpnsIpfsDist = "/ipns/dist.ipfs.io"

	// Distribution environ variable
	envIpfsDistPath = "IPFS_DIST_PATH"
)

type Fetcher interface {
	// Fetch attempts to fetch the file at the given ipfs path.
	// Returns io.ReadCloser on success, which caller must close.
	Fetch(ctx context.Context, filePath string) (io.ReadCloser, error)
	// SetDistPath sets the path to the distribution site for a Fetcher
	SetDistPath(distPath string)
}

// MultiFetcher holds multiple Fetchers and provides a Fetch that tries each
// until one succeeds.
type MultiFetcher struct {
	fetchers []Fetcher
}

type limitReadCloser struct {
	io.Reader
	io.Closer
}

// NewMultiFetcher creates a MultiFetcher with the given Fetchers.  The
// Fetchers are tried in order ther passed to this function.
func NewMultiFetcher(f ...Fetcher) Fetcher {
	mf := &MultiFetcher{
		fetchers: make([]Fetcher, len(f)),
	}
	copy(mf.fetchers, f)
	return mf
}

// Fetch attempts to fetch the file at each of its fetchers until one succeeds.
// Returns io.ReadCloser on success, which caller must close.
func (f *MultiFetcher) Fetch(ctx context.Context, ipfsPath string) (rc io.ReadCloser, err error) {
	for _, fetcher := range f.fetchers {
		rc, err = fetcher.Fetch(ctx, ipfsPath)
		if err == nil {
			// Transferred using this fetcher
			return
		}
	}
	return
}

// SetDistPath sets the path to the distribution site for all fetchers
func (f *MultiFetcher) SetDistPath(distPath string) {
	if !strings.HasPrefix(distPath, "/") {
		distPath = "/" + distPath
	}
	for _, fetcher := range f.fetchers {
		fetcher.SetDistPath(distPath)
	}
}

// NewLimitReadCloser returns a new io.ReadCloser with the reader wrappen in a
// io.LimitedReader limited to reading the amount specified.
func NewLimitReadCloser(rc io.ReadCloser, limit int64) io.ReadCloser {
	return limitReadCloser{
		Reader: io.LimitReader(rc, limit),
		Closer: rc,
	}
}

// GetDistPathEnv returns the IPFS path to the distribution site, using
// the value of environ variable specified by envIpfsDistPath.  If the environ
// variable is not set, then returns the provided distPath, and if that is not set
// then returns the IPNS path.
//
// To get the IPFS path of the latest distribution, if not overriddin by the
// environ variable: GetDistPathEnv(CurrentIpfsDist)
func GetDistPathEnv(distPath string) string {
	if dist := os.Getenv(envIpfsDistPath); dist != "" {
		return dist
	}
	if distPath == "" {
		return IpnsIpfsDist
	}
	return distPath
}
