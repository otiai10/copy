//go:build windows || js || plan9

package copy

func preserveLtimes(src, dest string, opt Options) error {
	return nil // Unsupported
}
