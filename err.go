package doubanfm

// {"err":"wrong_version", "r":1}
type dfmError struct {
	R   int
	Err string
}

func (e dfmError) Error() string {
	return e.Err
}
