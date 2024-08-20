//go:build !test_cow

package copy

const isTestingCopyOnWrite = false

func CopyInTest(src, dest string, opts ...Options) error {
	return Copy(src, dest, opts...)
}
