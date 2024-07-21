package helloworld

import "testing"

func TestHello(t *testing.T) {

	t.Run("saying hello to people", func(t *testing.T) {
		user_name := "Chris"

		got := Hello(user_name, "Spanish")
		want := "Hola, Chris"

		assertCorrectMessage(t, got, want)
	})

	t.Run("say, 'Hello, World' when an empty string is supplied", func(t *testing.T) {
		empty_string := ""
		got := Hello(empty_string, "")
		want := "Hello, World"

		assertCorrectMessage(t, got, want)
	})
}

func assertCorrectMessage(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
