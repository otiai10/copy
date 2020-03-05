package copy

// Options specifies optional actions on copying.
type Options struct {
	// OnSymlink can specify what to do on symlink
	OnSymlink func(p string) SymlinkAction
}

// SymlinkAction represents what to do on symlink.
type SymlinkAction int

const (
	// Deep creates hard-copy of contents.
	Deep SymlinkAction = iota
	// Shallow creates new symlink to the dest of symlink.
	Shallow
	// Skip does nothing with symlink.
	Skip
)

// DefaultOptions by default.
var DefaultOptions = Options{
	OnSymlink: func(string) SymlinkAction {
		return Shallow
	},
}
