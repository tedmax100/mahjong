package version

import (
	"fmt"
	"runtime"
)

// 這些變數會在編譯時透過 -ldflags 注入
var (
	// GitCommit is the git SHA at build time
	GitCommit = "unknown"
	// Version is the semantic version
	Version = "dev"
	// BuildTime is the build timestamp
	BuildTime = "unknown"
)

// Info returns formatted version information
func Info() string {
	return fmt.Sprintf("Version: %s, Commit: %s, Built: %s, Go: %s",
		Version, GitCommit, BuildTime, runtime.Version())
}

// Print logs the version information
func Print(serviceName string) {
	fmt.Println("")
	fmt.Printf("📦 %s Build Info\n", serviceName)
	fmt.Println("=====================================")
	fmt.Printf("   Version:  %s\n", Version)
	fmt.Printf("   Commit:   %s\n", GitCommit)
	fmt.Printf("   Built:    %s\n", BuildTime)
	fmt.Printf("   Go:       %s\n", runtime.Version())
	fmt.Println("=====================================")
	fmt.Println("")
}
