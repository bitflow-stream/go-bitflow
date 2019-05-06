package bitflow

type fileVanishChecker struct {
}

func (f *fileVanishChecker) setCurrentFile(name string) error {
	// TODO implement file check on Windows
	return nil
}

func (f *fileVanishChecker) hasFileVanished(name string) bool {
	// TODO implement file check on Windows
	return false
}

func IsFileClosedError(err error) bool {
	// TODO implement error check on Windows
	return false
}

func IsBrokenPipeError(err error) bool {
	// TODO implement error check on Windows
	return false
}
