package cdp

import "encoding/json"

// Target represents a Chrome DevTools Protocol debugging target.
type Target struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// Message is a CDP protocol frame (request or response or event).
type Message struct {
	ID     int             `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *CDPError       `json:"error,omitempty"`
}

// CDPError is the error payload in a CDP response.
type CDPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// EvaluateParams holds parameters for Runtime.evaluate.
type EvaluateParams struct {
	Expression    string `json:"expression"`
	ReturnByValue bool   `json:"returnByValue"`
	AwaitPromise  bool   `json:"awaitPromise,omitempty"`
	Timeout       int64  `json:"timeout,omitempty"`
}

// RemoteObject is a CDP Runtime.RemoteObject.
type RemoteObject struct {
	Type        string          `json:"type"`
	Subtype     string          `json:"subtype,omitempty"`
	Value       json.RawMessage `json:"value,omitempty"`
	Description string          `json:"description,omitempty"`
}

// EvaluateResult is the result of a Runtime.evaluate call.
type EvaluateResult struct {
	Result           RemoteObject      `json:"result"`
	ExceptionDetails *ExceptionDetails `json:"exceptionDetails,omitempty"`
}

// ExceptionDetails holds JS exception info from Runtime.evaluate.
type ExceptionDetails struct {
	Text      string        `json:"text"`
	Exception *RemoteObject `json:"exception,omitempty"`
}
