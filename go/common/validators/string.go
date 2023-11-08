package validators

func Length(s string, min, max int, msg string) (string, bool) {
	l := len(s)
	if l < min || l > max {
		return msg, false
	}

	return "", true
}
