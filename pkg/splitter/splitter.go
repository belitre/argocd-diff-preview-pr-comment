package splitter

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/belitre/argocd-diff-preview-pr-comment/pkg/logger"
)

// SplitResult represents a split part of the diff
type SplitResult struct {
	PartNumber int
	TotalParts int
	Content    string
	Size       int
}

// SplitDiffFile splits a markdown diff file if it exceeds the max length
// Returns the list of split results
func SplitDiffFile(inputPath string, maxLength int) ([]SplitResult, error) {
	log := logger.GetLogger()

	// Read the entire file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	// If the file is within the limit, return the original path
	if len(content) <= maxLength {
		log.Infof("File size (%d bytes) is within the limit (%d bytes). No splitting needed.", len(content), maxLength)
		return []SplitResult{
			{
				PartNumber: 1,
				TotalParts: 1,
				Content:    string(content),
				Size:       len(content),
			},
		}, nil
	}

	log.Infof("File size (%d bytes) exceeds limit (%d bytes). Splitting file...", len(content), maxLength)

	// Parse the file structure
	lines := strings.Split(string(content), "\n")

	// Find the header (everything before first <details>)
	headerEndIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "<details>") {
			headerEndIdx = i - 1
			break
		}
	}

	if headerEndIdx == -1 {
		return nil, fmt.Errorf("could not find <details> tag in the file")
	}

	header := strings.Join(lines[:headerEndIdx+1], "\n") + "\n"

	// Find the footer (stats at the end)
	footerStartIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "_Stats_:") {
			footerStartIdx = i
			break
		}
	}

	footer := ""
	if footerStartIdx != -1 {
		footer = "\n" + strings.Join(lines[footerStartIdx:], "\n")
	}

	// Reserve space for part indicator (e.g., "\n\n---\n**Part 999 of 999**\n")
	// This accounts for the maximum possible part numbers (3 digits each)
	// Total: 2 newlines + 3 chars (---) + 1 newline + 2 stars + "Part " (5) +
	//        3 digits + " of " (4) + 3 digits + 2 stars + 1 newline = ~25 bytes
	// Use 35 to be extra safe
	partIndicatorMaxSize := 35

	// Split the content into chunks
	// For the first chunk: maxLength - header - footer - partIndicator
	// For subsequent chunks: maxLength - partIndicator
	// We use the more restrictive first chunk size for all chunks to keep logic simple
	effectiveMaxLength := maxLength - len(header) - len(footer) - partIndicatorMaxSize
	if effectiveMaxLength < 100 {
		return nil, fmt.Errorf("max length too small to split file (effective content space: %d bytes)", effectiveMaxLength)
	}

	chunks := splitIntoChunks(lines, headerEndIdx+1, footerStartIdx, effectiveMaxLength)

	if len(chunks) == 0 {
		return nil, fmt.Errorf("failed to split file into valid chunks")
	}

	totalParts := len(chunks)

	// Build split results
	results := make([]SplitResult, 0, len(chunks))

	for i, chunk := range chunks {
		partIndicator := ""
		if totalParts > 1 {
			partIndicator = fmt.Sprintf("\n\n---\n**Part %d of %d**\n", i+1, totalParts)
		}

		var fileContent string
		if i == 0 {
			// First part: include header, footer, and part indicator
			fileContent = header + chunk + footer + partIndicator
		} else {
			// Subsequent parts: only chunk content and part indicator
			fileContent = chunk + partIndicator
		}

		// Verify size doesn't exceed max length
		if len(fileContent) > maxLength {
			log.Warnf("File part %d size (%d bytes) exceeds max length (%d bytes) by %d bytes",
				i+1, len(fileContent), maxLength, len(fileContent)-maxLength)
		}

		results = append(results, SplitResult{
			PartNumber: i + 1,
			TotalParts: totalParts,
			Content:    fileContent,
			Size:       len(fileContent),
		})
	}

	return results, nil
}

// splitIntoChunks splits the content between start and end indices into chunks
// maxChunkSize is the maximum size for each chunk's content (not including header/footer)
func splitIntoChunks(lines []string, startIdx, endIdx, maxChunkSize int) []string {
	log := logger.GetLogger()

	if endIdx == -1 {
		endIdx = len(lines)
	}

	chunks := make([]string, 0)
	currentChunk := make([]string, 0)
	currentSize := 0

	// Track if we're inside an application's <details> block
	insideDetails := false
	currentApp := ""

	for i := startIdx; i < endIdx; i++ {
		line := lines[i]

		// Check if we're starting a new application details block
		if strings.Contains(line, "<details>") {
			insideDetails = true
			// Extract application name from the next line (summary)
			if i+1 < endIdx && strings.Contains(lines[i+1], "<summary>") {
				currentApp = extractAppName(lines[i+1])
			}
		}

		lineSize := len(line) + 1 // +1 for newline

		// Reserve space for closing tags if we're inside a details block
		closingTagsSize := 0
		if insideDetails {
			closingTagsSize = len("\n```\n\n</details>\n")
		}

		// Check if adding this line would exceed the limit
		if currentSize+lineSize+closingTagsSize > maxChunkSize && len(currentChunk) > 0 {
			// Close the current details block if we're inside one
			chunkContent := strings.Join(currentChunk, "\n")
			if insideDetails {
				chunkContent += "\n```\n\n</details>\n"
			}
			chunks = append(chunks, chunkContent)
			log.Debugf("Created chunk %d with size %d", len(chunks)-1, len(chunkContent))

			// Start new chunk with continuation notice if we're splitting an app
			currentChunk = make([]string, 0)
			if insideDetails && currentApp != "" {
				continuationHeader := fmt.Sprintf("<details>\n<summary>%s (continuation...)</summary>\n<br>\n\n```diff", currentApp)
				currentChunk = append(currentChunk, continuationHeader)
				currentSize = len(continuationHeader) + 1
			} else {
				currentSize = 0
			}
		}

		currentChunk = append(currentChunk, line)
		currentSize += lineSize

		// Check if we're closing a details block
		if strings.Contains(line, "</details>") {
			insideDetails = false
			currentApp = ""
		}
	}

	// Add the last chunk
	if len(currentChunk) > 0 {
		chunkContent := strings.Join(currentChunk, "\n")
		chunks = append(chunks, chunkContent)
		log.Debugf("Created final chunk %d with size %d", len(chunks)-1, len(chunkContent))
	}

	return chunks
}

// extractAppName extracts the application name from a summary line
func extractAppName(summaryLine string) string {
	// Format: <summary>app-name (path)</summary>
	start := strings.Index(summaryLine, "<summary>")
	if start == -1 {
		return ""
	}
	start += len("<summary>")

	end := strings.Index(summaryLine, "</summary>")
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(summaryLine[start:end])
}

// CountFileSize returns the size of a file in bytes
func CountFileSize(filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}
	return len(content), nil
}

// ReadFileLines reads a file and returns it as a slice of lines
func ReadFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
