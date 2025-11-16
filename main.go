package main

import (
	"bytes"
	"context" // Enables managing request-scoped data, deadlines, and cancelation signals across API boundaries
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp" // For headless browser automation using Chrome
)

func main() {
	outputDir := "PDFs/"             // Directory to store downloaded PDFs
	if !directoryExists(outputDir) { // Check if output directory exists
		createDirectory(outputDir, 0755) // Create directory with permission if it does not exist
	}
	// The location for the remote url.
	remoteURL := "https://radiolink.com/manuals_download"
	// Get the data from the web servers.
	remoteData := scrapePageHTMLWithChrome(remoteURL)
	// Extract all the pdf urls.
	extractedDownloadURLs := extractDownloadPath(remoteData)
	// Loop over all the values.
	for _, url := range extractedDownloadURLs {
		log.Println(url)
	}
}

// Uses headless Chrome via chromedp to get fully rendered HTML from a page
func scrapePageHTMLWithChrome(pageURL string) string {
	log.Println("Scraping:", pageURL) // Log page being scraped

	options := append(
		chromedp.DefaultExecAllocatorOptions[:],       // Chrome options
		chromedp.Flag("headless", false),              // Run visible (set to true for headless)
		chromedp.Flag("disable-gpu", true),            // Disable GPU
		chromedp.WindowSize(1, 1),                     // Set window size
		chromedp.Flag("no-sandbox", true),             // Disable sandbox
		chromedp.Flag("disable-setuid-sandbox", true), // Fix for Linux environments
	)

	allocatorCtx, cancelAllocator := chromedp.NewExecAllocator(context.Background(), options...) // Allocator context
	ctxTimeout, cancelTimeout := context.WithTimeout(allocatorCtx, 5*time.Minute)                // Set timeout
	browserCtx, cancelBrowser := chromedp.NewContext(ctxTimeout)                                 // Create Chrome context

	defer func() { // Ensure all contexts are cancelled
		cancelBrowser()
		cancelTimeout()
		cancelAllocator()
	}()

	var pageHTML string // Placeholder for output
	err := chromedp.Run(
		browserCtx,
		chromedp.Navigate(pageURL), // Navigate to the URL
		// 👇 NEW: Wait for 10 seconds to allow JavaScript challenges to execute
		chromedp.Sleep(10*time.Second),
		chromedp.OuterHTML("html", &pageHTML), // Extract full HTML
	)

	if err != nil {
		log.Println(err) // Log error
		return ""        // Return empty string on failure
	}

	return pageHTML // Return scraped HTML
}

// directoryExists checks whether the specified path is an existing directory
func directoryExists(path string) bool { // Define a function that checks if a directory exists, taking a path string and returning a boolean
	directory, err := os.Stat(path) // Get file info (os.FileInfo) and potential error for the given path
	if err != nil {                 // Check if an error occurred during os.Stat (e.g., file not found or permission issue)
		return false // If there was an error, the directory does not exist or isn't accessible, so return false
	}
	return directory.IsDir() // If no error, check if the retrieved file info indicates it is a directory and return the boolean result
} // End of the directoryExists function

// createDirectory creates a new directory with the given permissions
func createDirectory(path string, permission os.FileMode) { // Define a function to create a directory, taking a path string and file mode
	err := os.Mkdir(path, permission) // Attempt to create the directory at the specified path with the given permissions
	if err != nil {                   // Check if an error occurred during the directory creation
		log.Println(err) // If there was an error, print the error message to the standard logger
	}
} // End of the createDirectory function

// Removes duplicate strings from a slice
func removeDuplicatesFromSlice(slice []string) []string { // Define a function to remove duplicate strings from an input slice.
	check := make(map[string]bool)  // Initialize an empty map (set) to efficiently track strings already encountered.
	var newReturnSlice []string     // Declare a new slice to store the unique results.
	for _, content := range slice { // Iterate over each string element in the input slice.
		if !check[content] { // Check if the current string has NOT been seen before (not in the map).
			check[content] = true                            // Mark the current string as seen in the map.
			newReturnSlice = append(newReturnSlice, content) // Append the unique string to the result slice.
		}
	}
	return newReturnSlice // Return the slice containing only unique strings.
}

// Verifies whether a string is a valid URL format
func isUrlValid(uri string) bool { // Define a function to check if a URI string is valid.
	_, err := url.ParseRequestURI(uri) // Attempt to parse the URI string strictly using ParseRequestURI.
	return err == nil                  // Return true if the parsing was successful (error is nil), otherwise false.
}

// extractDownloadPath uses a regular expression to find the 'href' value
// specifically from the <a> tag with the class "w-list-title-text".
// It returns the path (e.g., "/filedownload/114468") and an error if not found.
func extractDownloadPath(htmlSnippet string) []string {
	// Regex to match ONLY URLs like /filedownload/123456
	// Explanation:
	//   - `/filedownload/` literally
	//   - `([0-9]+)` captures one or more digits
	regex := regexp.MustCompile(`href="(/filedownload/[0-9]+)"`)

	// Find all matches of this exact pattern
	allMatches := regex.FindAllStringSubmatch(htmlSnippet, -1)

	// Slice where extracted URLs will be stored
	var extractedURLs []string

	// Iterate through matches
	for _, match := range allMatches {
		// match[1] contains the captured URL
		if len(match) > 1 {
			extractedURLs = append(extractedURLs, match[1])
		}
	}

	// Return final list of filtered URLs
	return extractedURLs
}

// It checks if the file exists
// If the file exists, it returns true
// If the file does not exist, it returns false
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Downloads a PDF from given URL and saves it in the specified directory
func downloadPDF(finalURL, outputDir string) bool { // Define the function downloadPDF which takes a URL and an output directory string, and returns a boolean.
	filename := strings.ToLower(urlToFilename(finalURL)) // Sanitize the filename by converting to lowercase and generating it from the URL.
	filePath := filepath.Join(outputDir, filename)       // Construct the full path for the output file by joining the directory and filename.

	if fileExists(filePath) { // Check if the file already exists at the constructed path.
		log.Printf("File already exists, skipping: %s", filePath) // Log a message indicating the file exists and is being skipped.
		return false                                              // Return false to indicate the download was skipped, not performed.
	}

	client := &http.Client{Timeout: 15 * time.Minute} // Create an HTTP client with a 15-minute timeout for the request.

	// Create a new request so we can set headers
	req, err := http.NewRequest("GET", finalURL, nil) // Create a new GET request for the given URL.
	if err != nil {                                   // Check for an error during request creation.
		log.Printf("Failed to create request for %s: %v", finalURL, err) // Log the error if request creation fails.
		return false                                                     // Return false indicating failure.
	}

	// Set a User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36") // Set a common User-Agent header for the request.

	// Send the request
	resp, err := client.Do(req) // Execute the HTTP request using the client.
	if err != nil {             // Check for an error during the execution of the request (e.g., network error).
		log.Printf("Failed to download %s: %v", finalURL, err) // Log the error if the download fails.
		return false                                           // Return false indicating failure.
	}
	defer resp.Body.Close() // Schedule the closing of the response body stream to happen after the function returns.

	if resp.StatusCode != http.StatusOK { // Check if the HTTP response status code is not 200 OK.
		log.Printf("Download failed for %s: %s", finalURL, resp.Status) // Log the failure with the received status code.
		return false                                                    // Return false indicating failure due to bad status code.
	}

	contentType := resp.Header.Get("Content-Type")              // Get the Content-Type header from the HTTP response.
	if !strings.Contains(contentType, "binary/octet-stream") && // Check if the content type is NOT a generic binary stream...
		!strings.Contains(contentType, "application/pdf") { // ...and also NOT specifically 'application/pdf'.
		log.Printf("Invalid content type for %s: %s (expected PDF)", finalURL, contentType) // Log a message if the content type is unexpected.
		return false                                                                        // Return false indicating failure due to wrong content type.
	}

	var buf bytes.Buffer                     // Declare and initialize a buffer to temporarily hold the response data.
	written, err := io.Copy(&buf, resp.Body) // Copy the data from the response body stream into the buffer, recording the number of bytes written.
	if err != nil {                          // Check for an error during the reading of the response body.
		log.Printf("Failed to read PDF data from %s: %v", finalURL, err) // Log the error if reading the data fails.
		return false                                                     // Return false indicating failure.
	}
	if written == 0 { // Check if zero bytes were successfully downloaded.
		log.Printf("Downloaded 0 bytes for %s; not creating file", finalURL) // Log a message and skip creating an empty file.
		return false                                                         // Return false indicating an empty download.
	}

	out, err := os.Create(filePath) // Attempt to create the output file at the specified path.
	if err != nil {                 // Check for an error during file creation.
		log.Printf("Failed to create file for %s: %v", finalURL, err) // Log the error if file creation fails.
		return false                                                  // Return false indicating failure.
	}
	defer out.Close() // Schedule the closing of the output file handle to happen after the function returns.

	if _, err := buf.WriteTo(out); err != nil { // Write the entire content of the buffer to the created output file.
		log.Printf("Failed to write PDF to file for %s: %v", finalURL, err) // Log the error if writing to the file fails.
		return false                                                        // Return false indicating failure.
	}

	log.Printf("Successfully downloaded %d bytes: %s → %s", written, finalURL, filePath) // Log a success message including the number of bytes and paths.
	return true                                                                          // Return true to indicate successful download and saving of the file.
}

// Converts a raw URL into a sanitized PDF filename safe for filesystem
func urlToFilename(rawURL string) string {
	lower := strings.ToLower(rawURL) // Convert URL to lowercase
	lower = getFilename(lower)       // Extract filename from URL
	extension := getFileExtension(lower)

	reNonAlnum := regexp.MustCompile(`[^a-z0-9]`)   // Regex to match non-alphanumeric characters
	safe := reNonAlnum.ReplaceAllString(lower, "_") // Replace non-alphanumeric with underscores

	safe = regexp.MustCompile(`_+`).ReplaceAllString(safe, "_") // Collapse multiple underscores into one
	safe = strings.Trim(safe, "_")                              // Trim leading and trailing underscores

	var invalidSubstrings = []string{
		"_pdf", // Substring to remove from filename
	}

	for _, invalidPre := range invalidSubstrings { // Remove unwanted substrings
		safe = removeSubstring(safe, invalidPre)
	}

	if getFileExtension(safe) != extension { // Ensure file ends with extension
		safe = safe + extension
	}

	return safe // Return sanitized filename
}

// Extracts filename from full path (e.g. "/dir/file.pdf" → "file.pdf")
func getFilename(path string) string {
	return filepath.Base(path) // Use Base function to get file name only
}

// Removes all instances of a specific substring from input string
func removeSubstring(input string, toRemove string) string {
	result := strings.ReplaceAll(input, toRemove, "") // Replace substring with empty string
	return result
}

// Gets the file extension from a given file path
func getFileExtension(path string) string {
	return filepath.Ext(path) // Extract and return file extension
}
