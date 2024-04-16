package routes

import "net/http"

// BasicHandler contains just a response, that is used to show the templating
func BasicHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello there"))
}
