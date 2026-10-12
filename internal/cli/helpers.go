package cli

func orUnknown(s string) string {
	if s == "" {
		return "(unknown)"
	}
	return s
}
