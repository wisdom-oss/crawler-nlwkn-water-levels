package routes

import (
	"errors"
	"net/http"

	wisdomMiddleware "github.com/wisdom-oss/microservice-middlewares/v4"
)

func BasicWithErrorHandling(w http.ResponseWriter, r *http.Request) {
	// access the error handlers
	errorHandler := r.Context().Value(wisdomMiddleware.ErrorChannelName).(chan<- interface{})
	statusChannel := r.Context().Value(wisdomMiddleware.StatusChannelName).(<-chan bool)
	// now publish an error to each of the wisdom errors
	errorHandler <- errors.New("native test error")
	// now wait for the error to be handled
	<-statusChannel
	return
}
