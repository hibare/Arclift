package version

import (
	"github.com/hibare/GoCommon/v2/pkg/version"
	"github.com/hibare/GoS3Backup/internal/constants"
)

var (
	CurrentVersion = "0.0.0"
	V              = version.NewVersion(constants.GithubOwner, constants.ProgramPrettyIdentifier, CurrentVersion, version.Options{})
)
