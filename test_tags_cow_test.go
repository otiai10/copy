//go:build test_cow

package copy

// For this test, we want to run all the tests with CopyOnWritePreferred.
//
// This ensures that everything still functions as expected when CopyOnWrite
// is enabled.
const isTestingCopyOnWrite = true

func CopyInTest(src, dest string, opts ...Options) error {
	opt := assureOptions(src, dest, opts...)
	opt.CopyOnWrite = CopyOnWritePreferred
	return Copy(src, dest, opt)
}
