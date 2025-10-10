package appinfo

// AppName and Version are used to identify this application to the C library
// and for logging/version output. Override Version via -ldflags at build time.
// Example (Windows cmd):
//   go build -ldflags "-X 'win-sound-dev-go-bridge/pkg/appinfo.Version=1.0.0'" .

var (
	// AppName is passed to the underlying DLL.
	AppName = "win-sound-dev-go-bridge"
	// Version can be injected at build time using -ldflags -X
	Version = "dev"
)
