package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type HTTPRequest struct {
	Name    string
	Method  string
	URL     string
	Headers string
	Body    string
	Params  string
}

type HTTPFileData struct {
	Requests   []HTTPRequest
	GlobalVars map[string]string
}

// Serialize HTTPFileData to .http file format
func (h *HTTPFileData) ToHTTPFileFormat() string {
	var sb strings.Builder
	sb.WriteString("### ||| POSTBEAR |||\n")
	// Write global variables
	if len(h.GlobalVars) > 0 {
		sb.WriteString("### Global Variables\n")
		for k, v := range h.GlobalVars {
			sb.WriteString(fmt.Sprintf("@%s = %s\n", k, v))
		}
		sb.WriteString("\n")
	}
	// Write requests
	for _, req := range h.Requests {
		sb.WriteString(fmt.Sprintf("### %s\n", req.Name))
		sb.WriteString(fmt.Sprintf("%s %s\n", req.Method, req.URL))
		// Write headers as HeaderName: Value per line
		if req.Headers != "" {
			headers, _ := parseHeadersToMap(req.Headers)
			for hk, hv := range headers {
				sb.WriteString(fmt.Sprintf("%s: %s\n", hk, hv))
			}
		}
		if req.Body != "" {
			sb.WriteString("\n" + req.Body + "\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// parseHeadersToMap parses a JSON or Go map string to a map[string]string
func parseHeadersToMap(headers string) (map[string]string, error) {
	// We'll decode the JSON into a map[string]interface{}.
	// This is a flexible approach that can handle various JSON value types.
	var result map[string]interface{}

	// Unmarshal the JSON string into the result map.
	if err := json.Unmarshal([]byte(headers), &result); err != nil {
		return nil, err
	}

	// Create a new map[string]string to store the converted data.
	stringMap := make(map[string]string)

	// Iterate over the decoded map.
	for key, value := range result {
		// Convert the value to a string.
		// The fmt.Sprintf function is a convenient way to do this.
		// It will handle different types (like numbers, booleans) gracefully.
		stringValue := fmt.Sprintf("%v", value)
		stringMap[key] = stringValue
	}

	return stringMap, nil
}

// Save HTTPFileData to a .http file in the current working directory
func SaveHTTPFile(data *HTTPFileData, filename string) error {
	if !strings.HasSuffix(filename, ".http") {
		filename += ".http"
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(cwd, filename)
	content := data.ToHTTPFileFormat()
	return os.WriteFile(path, []byte(content), 0644)
}

// LoadGlobalVarsFromHTTPFile loads global variables from the postbear.http file
func LoadGlobalVarsFromHTTPFile(filename string) map[string]string {
	vars := make(map[string]string)
	cwd, err := os.Getwd()
	if err != nil {
		return vars
	}
	path := filepath.Join(cwd, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return vars
	}
	lines := strings.Split(string(content), "\n")
	inGlobals := false
	for _, line := range lines {
		if strings.HasPrefix(line, "### Global Variables") {
			inGlobals = true
			continue
		}
		if inGlobals {
			if strings.HasPrefix(line, "### ") && !strings.HasPrefix(line, "### Global Variables") {
				break
			}
			if strings.HasPrefix(line, "@") {
				parts := strings.SplitN(line[1:], " = ", 2)
				if len(parts) == 2 {
					vars[parts[0]] = parts[1]
				}
			}
		}
	}
	return vars
}

// LoadHTTPFile loads the HTTPFileData (requests and global vars) from a .http file
func LoadHTTPFile(filename string) (*HTTPFileData, error) {
	data := &HTTPFileData{
		Requests:   []HTTPRequest{},
		GlobalVars: map[string]string{},
	}
	cwd, err := os.Getwd()
	if err != nil {
		return data, err
	}
	path := filepath.Join(cwd, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return data, err
	}
	lines := strings.Split(string(content), "\n")
	inGlobals := false
	inRequest := false
	var req HTTPRequest
	headerLines := []string{}
	processingHeaders := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for section headers
		if strings.HasPrefix(trimmedLine, "### Global Variables") {
			inGlobals = true
			inRequest = false
			processingHeaders = false
			continue
		}
		if strings.HasPrefix(trimmedLine, "### ") && !strings.HasPrefix(trimmedLine, "### Global Variables") {
			// A new request is starting, so process the previous one if it exists
			if inRequest {
				// We've finished processing headers for the previous request
				if processingHeaders {
					// Manually create a JSON string from header lines. This is fragile.
					jsonString, err := headerLinesToJSON(headerLines)
					if err == nil {
						// Pass the generated JSON string to the user-provided function
						headersMap, err := parseHeadersToMap(jsonString)
						if err == nil {
							// Marshal the map back into a JSON string for req.Headers
							b, err := json.MarshalIndent(headersMap, "", "  ")
							if err == nil {
								req.Headers = string(b)
							}
						}
					}
					headerLines = []string{}
					processingHeaders = false
				}
				data.Requests = append(data.Requests, req)
			}

			// Reset for the new request
			req = HTTPRequest{}
			inGlobals = false
			inRequest = true
			req.Name = strings.TrimPrefix(trimmedLine, "### ")
			continue
		}

		// Logic for parsing global variables
		if inGlobals {
			if strings.HasPrefix(trimmedLine, "@") {
				parts := strings.SplitN(trimmedLine[1:], " = ", 2)
				if len(parts) == 2 {
					data.GlobalVars[parts[0]] = parts[1]
				}
			}
			continue
		}

		// Logic for parsing requests
		if inRequest {
			// First line of the request (Method and URL)
			if req.Method == "" && req.URL == "" && trimmedLine != "" {
				parts := strings.SplitN(trimmedLine, " ", 2)
				if len(parts) == 2 {
					req.Method = parts[0]
					req.URL = parts[1]
				}
				continue
			}

			// Logic for parsing headers
			if req.Method != "" && req.URL != "" && !processingHeaders && strings.Contains(line, ":") && !strings.HasPrefix(trimmedLine, "{") && !strings.HasPrefix(trimmedLine, "[") && trimmedLine != "" {
				processingHeaders = true
				headerLines = append(headerLines, trimmedLine)
				continue
			}

			// Continue collecting header lines
			if processingHeaders {
				if trimmedLine == "" {
					// End of headers, now process them
					jsonString, err := headerLinesToJSON(headerLines)
					if err == nil {
						headersMap, err := parseHeadersToMap(jsonString)
						if err == nil {
							b, err := json.MarshalIndent(headersMap, "", "  ")
							if err == nil {
								req.Headers = string(b)
							}
						}
					}
					headerLines = []string{}
					processingHeaders = false
					continue
				} else if strings.HasPrefix(trimmedLine, "{") || strings.HasPrefix(trimmedLine, "[") {
					// The body has started, so stop collecting headers
					jsonString, err := headerLinesToJSON(headerLines)
					if err == nil {
						headersMap, err := parseHeadersToMap(jsonString)
						if err == nil {
							b, err := json.MarshalIndent(headersMap, "", "  ")
							if err == nil {
								req.Headers = string(b)
							}
						}
					}
					headerLines = []string{}
					processingHeaders = false
					req.Body += line + "\n"
					continue
				} else {
					headerLines = append(headerLines, trimmedLine)
					continue
				}
			}

			// Logic for parsing body
			if strings.HasPrefix(trimmedLine, "{") || strings.HasPrefix(trimmedLine, "[") {
				req.Body += line + "\n"
				continue
			}
		}
	}

	// Finalize the last request if there is one
	if inRequest && (req.Method != "" || req.URL != "") {
		if processingHeaders {
			jsonString, err := headerLinesToJSON(headerLines)
			if err == nil {
				headersMap, err := parseHeadersToMap(jsonString)
				if err == nil {
					b, err := json.MarshalIndent(headersMap, "", "  ")
					if err == nil {
						req.Headers = string(b)
					}
				}
			}
		}
		data.Requests = append(data.Requests, req)
	}

	// Clean up body by trimming trailing newline
	for i, r := range data.Requests {
		data.Requests[i].Body = strings.TrimSpace(r.Body)
	}

	return data, nil
}

// headerLinesToJSON is a helper function to manually construct a JSON string
// from the `Header: Value` lines. This is a hacky workaround to use the
// provided `parseHeadersToMap` function, which expects a single JSON string.
func headerLinesToJSON(lines []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("{")
	for i, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf(`"%s": "%s"`, key, value))
		}
	}
	sb.WriteString("}")
	return sb.String(), nil
}
