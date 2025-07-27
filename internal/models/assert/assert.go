package assert

import "testing"

// проверяет, одинаковы ли значения, и в случае их отличия выбрасывает ошибку
func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}
