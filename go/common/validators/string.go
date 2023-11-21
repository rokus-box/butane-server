package validators

func Length(s string, min, max int) bool {
	l := len(s)
	if l < min || l > max {
		return false
	}

	return true
}
