package version

//
// Must be set at build via
// -ldflags "-X ./pkg/version.Version=`cat VERSION`"
// or
// -ldflags "-X ./pkg/version.Version=`git describe --tags`"
//

// app - info
const (
	Version string = "0.0.1"
)

// build - info
const (
	BuildVersion string = "2015.6.2-6-gfd7e2d1-dev"
	BuildTime    string = "2015-06-16-0431 UTC"
	BuildCount   string = ""
	BuildUnix    string = ""
)

// vcs - branch
const (
	BranchName string = ""
	RepoURI    string = ""
	Author     string = ""
)

// vcs - commits
const (
	CommitHash     string = ""
	CommitID       string = ""
	CommitTime     string = ""
	CommitTimeUnix string = ""
	CommitTimePrev string = ""
)
