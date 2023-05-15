package version

func GetVersionString(includeCommit bool) string {
	ver := Version

	if GitDescribe != "" {
		ver = GitDescribe
	}

	if includeCommit && GitCommit != "" {
		ver += " (" + GitCommit + ")"
	}

	return ver
}
