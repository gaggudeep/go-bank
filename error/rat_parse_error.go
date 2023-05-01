package error

type RatParseError struct {
	Msg string
}

func (err *RatParseError) Error() string {
	return err.Msg
}
