package weibo

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool { return &b }

// strPtr returns a pointer to a string value
func strPtr(s string) *string { return &s }

// intPtr returns a pointer to an int value
func intPtr(i int) *int { return &i }

// int64Ptr returns a pointer to an int64 value
func int64Ptr(i int64) *int64 { return &i }

// ptrConnectionState returns a pointer to a ConnectionState value
func ptrConnectionState(s ConnectionState) *ConnectionState { return &s }

// ifThenElse returns a or b based on cond
func ifThenElse(cond bool, a, b ConnectionState) ConnectionState {
	if cond {
		return a
	}
	return b
}

// derefBool dereferences a bool pointer safely
func derefBool(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}

// derefString dereferences a string pointer safely
func derefString(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

// derefInt dereferences an int pointer safely
func derefInt(i *int, defaultVal int) int {
	if i == nil {
		return defaultVal
	}
	return *i
}

// derefInt64 dereferences an int64 pointer safely
func derefInt64(i *int64, defaultVal int64) int64 {
	if i == nil {
		return defaultVal
	}
	return *i
}
