package profile

import "fmt"

// Loader identities are pinned release artifacts, not user-selectable DLLs.
// Keep old identities when supporting a future release: journals are immutable.
const (
	LoaderID            = "asi-loader"
	LoaderVersion       = "9.7.4"
	LoaderSHA256        = "031a3e5576d91dce1e438d36b9a3d462c7334ab4791990a8ff1e3ddc0e132daf"
	LoaderBytes   int64 = 1198304
)

var LoaderFiles = []Fingerprint{
	{Path: "wininet.dll", SHA256: LoaderSHA256},
}

func LoaderTarget(path string) error {
	for _, file := range LoaderFiles {
		if path == file.Path {
			return nil
		}
	}
	return fmt.Errorf("unsupported loader target: %q", path)
}

// AddedTarget is the union of supported absent-origin targets. Manifest
// validation separately restricts loader targets to the pinned schema-3 package.
func AddedTarget(path string) error {
	if PluginTarget(path) == nil || LoaderTarget(path) == nil {
		return nil
	}
	return fmt.Errorf("unsupported added target: %q", path)
}
