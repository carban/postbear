package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/charmbracelet/lipgloss"
)

func sendByTUI(m Model) (string, string, string) {
	variables := LoadGlobalVarsFromHTTPFile(m.filepath)
	method := strings.ToUpper(strings.TrimSpace(m.methodField.Value()))
	URL := strings.TrimSpace(m.urlField.Value())
	headersJSON := strings.TrimSpace(m.headersArea.Value())
	// paramsJSON := strings.TrimSpace(m.tabContent[paramsTab].Value())

	// Parse variables into a map
	// var variables map[string]string
	// if variablesJSON == "" { // "" is not valid string to unmarshal. occurs when user doesn't init any env vars.
	// 	variables = make(map[string]string)
	// } else {
	// 	er := json.Unmarshal([]byte(variablesJSON), &variables)
	// 	if er != nil {
	// 		return "\n Error parsing Env Variables", "Incorrect Env Variables", ""
	// 	}
	// }

	URL = replacePlaceholders(URL, variables)
	headersJSON = replacePlaceholders(headersJSON, variables)
	// paramsJSON = replacePlaceholders(paramsJSON, variables)

	// Parse JSON into a map
	var headers map[string]string
	err := json.Unmarshal([]byte(headersJSON), &headers)
	if err != nil {
		return " \n Error parsing Headers \n\n Correct the Headers format", " Incorrect Headers ", ""
	}

	// if paramsJSON != "" {
	// 	var params map[string]string
	// 	Err := json.Unmarshal([]byte(paramsJSON), &params)
	// 	if Err != nil {
	// 		return " \n Error parsing Params \n\n Correct the Params format", " Incorrect Params ", ""
	// 	}

	// 	// Create a URL object
	// 	parsedURL, err := url.Parse(URL)
	// 	if err != nil {
	// 		return " \n Error parsing Params \n\n Correct the Params format", " Incorrect Params ", ""
	// 	}

	// 	// Add query parameters to the URL
	// 	q := parsedURL.Query()
	// 	for key, value := range params {
	// 		q.Set(key, value)
	// 	}
	// 	parsedURL.RawQuery = q.Encode()

	// 	URL = parsedURL.String()
	// }

	// Prepare payload for methods that need it
	var payload io.Reader
	switch method {
	case "POST", "PUT", "PATCH":
		content := replacePlaceholders(m.bodyArea.Value(), variables)
		payload = bytes.NewBuffer([]byte(content))
	default:
		payload = nil
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, URL, payload)
	if err != nil {
		return "Failed to make request\n\n" + err.Error(), "", ""
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// --- Start the timer before sending the request ---
	startTime := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return "Failed to make request\n\n" + err.Error(), "", ""
	}
	defer resp.Body.Close()

	// --- Calculate the elapsed time after receiving the response headers ---
	duration := time.Since(startTime)
	ms := duration.Milliseconds()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Failed to read response body\n\n" + err.Error(), "", ""
	}
	return string(body), fmt.Sprint(resp.StatusCode), fmt.Sprintf(" %vms ", ms)
}

func SendByCLI(method string, url string, simpleOutput bool, payloadIdx int) {
	var reqBody io.Reader
	if (method == "POST" || method == "PUT" || method == "PATCH") && len(os.Args) > payloadIdx {
		payload := os.Args[payloadIdx]
		reqBody = bytes.NewBuffer([]byte(payload))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "my-simple-go-client/1.0")

	client := &http.Client{}
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	duration := time.Since(startTime)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	if simpleOutput {
		printResesponseBody(body)
		return
	}

	statusStyle := codes200Style

	if resp.StatusCode >= 500 {
		statusStyle = codes500Style
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		statusStyle = codes400Style
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		statusStyle = codes300Style
	}

	// --- Print all available response information using lipgloss ---
	urlStyle := boldStyle.Foreground(lipgloss.Color("5")).Background(lipgloss.Color("17")).Padding(0, 1)
	labelStyle := boldStyle.Foreground(lipgloss.Color("6"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	headerStyle := boldStyle.Foreground(lipgloss.Color("13"))

	methodStyle := otherMethodStyle
	switch method {
	case "GET":
		methodStyle = getMethodStyle
	case "POST":
		methodStyle = postMethodStyle
	case "PUT":
		methodStyle = putMethodStyle
	case "PATCH":
		methodStyle = patchMethodStyle
	case "DELETE":
		methodStyle = deleteMethodStyle
	case "INFO":
		methodStyle = infoMethodStyle
	}

	fmt.Println(methodStyle.Render(method) + urlStyle.Render(url))
	fmt.Println(statusStyle.Render(resp.Status))
	fmt.Println(labelStyle.Render("StatusCode:") + " " + valueStyle.Render(fmt.Sprintf("%d", resp.StatusCode)))
	fmt.Println(labelStyle.Render("Protocol:") + " " + valueStyle.Render(resp.Proto))
	fmt.Println(labelStyle.Render("ContentLength:") + " " + valueStyle.Render(fmt.Sprintf("%d", resp.ContentLength)))
	fmt.Println(labelStyle.Render("Response Time:") + " " + valueStyle.Render(fmt.Sprintf("%vms", duration.Milliseconds())))
	fmt.Println(headerStyle.Render("Headers:"))
	for key, values := range resp.Header {
		fmt.Println(labelStyle.Render("  "+key+":") + " " + valueStyle.Render(strings.Join(values, ", ")))
	}
	// --- Print the final endpoint result (response body) with colors ---
	fmt.Println(headerStyle.Render("Response:"))
	printResesponseBody(body)

}

func printResesponseBody(body []byte) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(body), &obj)
	fb := colorjson.NewFormatter()
	fb.Indent = 2
	s, _ := fb.Marshal(obj)
	fmt.Println(string(s))
}
