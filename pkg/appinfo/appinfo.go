package appinfo

var (
	// AppName is passed to the underlying DLL.
	AppName = "win-sound-scanner"
	// Version can be injected at build time using -ldflags -X
	Version = "dev"
)
