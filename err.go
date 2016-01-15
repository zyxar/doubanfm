package doubanfm

// {"err":"wrong_version", "r":1}
type dfmError struct {
	R       int
	Err     string
	Warning string
}

func (e dfmError) Error() string {
	if e.Err != "" {
		return e.Err
	}
	return e.Warning
}
