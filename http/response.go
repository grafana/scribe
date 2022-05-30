package swhttp

import (
	"fmt"
	"io"
	"net/http"
)

// HandleResponse is a utility function for standardizing failed responses. It checks the HTTP status code for a success response (200-299), and will attach the body of the response in the event of a non-200 response.
// This should be called immediately after an HTTP request, rather than checking immediately for the error.
func HandleResponse(res http.Response, err error) error {
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("non-200 response: %s. Error reading response body: %w", res.Status, err)
		}

		return fmt.Errorf("non-200 response: %s. body: '%s'", res.Status, string(b))
	}

	return nil
}
