package ssutil

func HasEmpty(items ...string) bool {
	for _, item := range items {
		if item == "" {
			return true
		}
	}
	return false
}

func FirstNonEmpty(items ...string) string {
	for _, item := range items {
		if item != "" {
			return item
		}
	}
	return ""
}
