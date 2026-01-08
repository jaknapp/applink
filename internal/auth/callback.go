package auth

import (
	"fmt"
	"net/http"
)

// startCallbackServer starts a local HTTP server to receive OAuth callbacks
func startCallbackServer(port int, expectedState string, codeChan chan<- string, errChan chan<- error) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check for error
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, errorHTML, errParam, errDesc)
			errChan <- fmt.Errorf("%s: %s", errParam, errDesc)
			return
		}

		// Verify state
		state := r.URL.Query().Get("state")
		if state != expectedState {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, errorHTML, "Invalid state", "State parameter mismatch")
			errChan <- fmt.Errorf("state mismatch: possible CSRF attack")
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, errorHTML, "Missing code", "No authorization code received")
			errChan <- fmt.Errorf("no authorization code in callback")
			return
		}

		// Success!
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, successHTML)
		codeChan <- code
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	return server
}

const successHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .card {
            background: white;
            padding: 3rem;
            border-radius: 1rem;
            box-shadow: 0 20px 40px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 400px;
        }
        .checkmark {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        h1 {
            color: #1a1a2e;
            margin: 0 0 0.5rem 0;
        }
        p {
            color: #666;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="checkmark">✓</div>
        <h1>Authentication Successful</h1>
        <p>You can close this window and return to the terminal.</p>
    </div>
</body>
</html>`

const errorHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Authentication Failed</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
        }
        .card {
            background: white;
            padding: 3rem;
            border-radius: 1rem;
            box-shadow: 0 20px 40px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 400px;
        }
        .error-icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        h1 {
            color: #1a1a2e;
            margin: 0 0 0.5rem 0;
        }
        p {
            color: #666;
            margin: 0;
        }
        .error-details {
            background: #f5f5f5;
            padding: 1rem;
            border-radius: 0.5rem;
            margin-top: 1rem;
            font-family: monospace;
            font-size: 0.9rem;
            color: #c0392b;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="error-icon">✗</div>
        <h1>Authentication Failed</h1>
        <p>Something went wrong during authentication.</p>
        <div class="error-details">%s: %s</div>
    </div>
</body>
</html>`
