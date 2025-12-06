package misc

import "errors"

// trim([]byte, int) ([]byte, error) - trims input byte slice to new byte slice of size n
func Trim(b []byte, n int) ([]byte, error) {
	if n > len(b) {
		return nil, errors.New("trimmed slice should have size smaller of or equal to origin slice")
	}
	result := make([]byte, n, n)
	for i := 0; i < n; i++ {
		result[i] = b[i]
	}
	return result, nil
}
