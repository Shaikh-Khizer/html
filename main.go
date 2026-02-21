package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const maxFileSize = 64 * 1024 // 64KB limit

// ----------------------
// Encoding Types
// ----------------------
type EncodingType int

const (
	Basic EncodingType = iota
	Full
	Smart
)

// ----------------------
// Basic Encoding
// ----------------------
func basicEncode(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}

func basicDecode(input string) string {
	replacer := strings.NewReplacer(
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&amp;", "&",
	)
	return replacer.Replace(input)
}

// ----------------------
// Full Encode/Decode
// ----------------------
func fullEncode(input string) string {
	var builder strings.Builder
	for _, r := range input {
		builder.WriteString(fmt.Sprintf("&#%d;", r))
	}
	return builder.String()
}

func fullDecode(input string) string {
	return html.UnescapeString(input)
}

// ----------------------
// Smart Encoding Detection
// ----------------------
func isEncoded(input string) bool {
	// Check for common HTML entities
	encodedPatterns := []string{
		"&amp;", "&lt;", "&gt;", "&quot;", "&#39;",
		"&#[0-9]+;", // Numeric entities
	}
	
	for _, pattern := range encodedPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	return false
}

func smartProcess(input string, encode bool, decode bool, encodingType EncodingType) string {
	if encodingType == Smart {
		if isEncoded(input) {
			// Input appears encoded, so decode
			if decode || (!encode && !decode) {
				return fullDecode(input)
			}
		} else {
			// Input appears unencoded, so encode
			if encode || (!encode && !decode) {
				return basicEncode(input)
			}
		}
		// If both flags are set, default to encode
		if encode && decode {
			return basicEncode(input)
		}
	}

	// Handle specific encoding types
	if encode {
		switch encodingType {
		case Full:
			return fullEncode(input)
		default:
			return basicEncode(input)
		}
	}
	
	if decode {
		switch encodingType {
		case Full:
			return fullDecode(input)
		default:
			return basicDecode(input)
		}
	}
	
	return input
}

// ----------------------
// File Utilities
// ----------------------
func readFileLimited(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file exceeds %dKB limit", maxFileSize/1024)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func writeFile(path string, content string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		fmt.Printf("File %s exists. Overwrite? (y/N): ", path)
		reader := bufio.NewReader(os.Stdin)
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))
		if resp != "y" && resp != "yes" {
			return errors.New("aborted by user")
		}
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// ----------------------
// Create HTML Structure
// ----------------------
func createHTML(title string, encodeTitle bool, encodingType EncodingType) string {
	if encodeTitle {
		switch encodingType {
		case Full:
			title = fullEncode(title)
		case Smart:
			if !isEncoded(title) {
				title = basicEncode(title)
			}
		default:
			title = basicEncode(title)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
</head>
<body>
    <h1>%s</h1>
    <p>Generated with HTML Tool</p>
</body>
</html>
`, title, title)
}

// ----------------------
// Filename Validation
// ----------------------
func isValidFilename(name string) bool {
	if name == "" {
		return false
	}
	invalid := regexp.MustCompile(`[<>:"/\\|?*]`)
	return !invalid.MatchString(name)
}

func ensureHTMLExt(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name + ".html"
	}
	if ext != ".html" && ext != ".htm" {
		return name + ".html"
	}
	return name
}

// ----------------------
// Input Detection
// ----------------------
func isHTMLContent(input string) bool {
	// Check for HTML tags
	htmlPatterns := []string{
		"<[a-z!]",  // Opening tags or <!
		"</[a-z]",   // Closing tags
		"<[A-Z]",    // Uppercase tags
	}
	
	for _, pattern := range htmlPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return true
		}
	}
	return false
}

func isNumeric(input string) bool {
	for _, r := range input {
		if !unicode.IsDigit(r) && r != ' ' && r != '\n' && r != '\t' {
			return false
		}
	}
	return len(strings.TrimSpace(input)) > 0
}

// ----------------------
// Main
// ----------------------
func main() {
	// Define flags
	encode := flag.Bool("e", false, "Encode HTML special characters")
	decode := flag.Bool("d", false, "Decode HTML entities")
	smart := flag.Bool("s", false, "Smart detection (auto-encode/decode based on content)")
	full := flag.Bool("full", false, "Use full numeric encoding/decoding")
	fileMode := flag.Bool("f", false, "Process input file")
	output := flag.String("o", "", "Output file")
	force := flag.Bool("force", false, "Force overwrite without confirmation")
	help := flag.Bool("h", false, "Show help")

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Determine encoding type
	encodingType := Basic
	if *full {
		encodingType = Full
	} else if *smart {
		encodingType = Smart
	}

	// Determine operation mode
	modeEncode := *encode
	modeDecode := *decode
	
	// If neither encode nor decode is specified, default based on smart mode
	if !modeEncode && !modeDecode {
		if encodingType == Smart {
			// In smart mode, we'll detect later
			modeEncode = true
			modeDecode = true
		} else {
			modeDecode = true // Default to decode for backward compatibility
		}
	}

	args := flag.Args()

	// Check for piped input
	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	// ----------------------
	// FILE MODE
	// ----------------------
	if *fileMode {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: Input file required")
			os.Exit(1)
		}

		content, err := readFileLimited(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		result := smartProcess(content, modeEncode, modeDecode, encodingType)

		if *output != "" {
			if err := writeFile(*output, result, *force); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Output written to %s\n", *output)
		} else {
			fmt.Print(result)
		}
		return
	}

	// ----------------------
	// PIPE MODE
	// ----------------------
	if isPiped {
		inputBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		input := string(inputBytes)

		result := smartProcess(input, modeEncode, modeDecode, encodingType)
		
		if *output != "" {
			if err := writeFile(*output, result, *force); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Output written to %s\n", *output)
		} else {
			fmt.Print(result)
		}
		return
	}

	// ----------------------
	// DIRECT ARGUMENT MODE
	// ----------------------
	if len(args) == 0 {
		// No arguments, enter interactive mode
		interactiveMode(modeEncode, modeDecode, encodingType, *output, *force)
		return
	}

	input := strings.Join(args, " ")

	// Check if input looks like a filename
	if isValidFilename(args[0]) && len(args) == 1 && !isHTMLContent(input) && !isNumeric(input) {
		// Likely a filename for HTML creation
		filename := ensureHTMLExt(args[0])
		title := strings.TrimSuffix(args[0], filepath.Ext(args[0]))
		content := createHTML(title, modeEncode, encodingType)

		if err := writeFile(filename, content, *force); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created HTML file: %s\n", filename)
		return
	}

	// Process as direct text
	result := smartProcess(input, modeEncode, modeDecode, encodingType)
	
	if *output != "" {
		if err := writeFile(*output, result, *force); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output written to %s\n", *output)
	} else {
		fmt.Print(result)
	}
}

func interactiveMode(encode bool, decode bool, encodingType EncodingType, outputFile string, force bool) {
	fmt.Println("Interactive Mode (Ctrl+D to exit)")
	fmt.Println("Enter text to process:")
	
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	
	input := strings.Join(lines, "\n")
	result := smartProcess(input, encode, decode, encodingType)
	
	if outputFile != "" {
		if err := writeFile(outputFile, result, force); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output written to %s\n", outputFile)
	} else {
		fmt.Println("\nProcessed output:")
		fmt.Println(result)
	}
}

func printHelp() {
	fmt.Println(`
HTML Encoder/Decoder Tool
==========================
Usage:
  html [OPTIONS] <text|filename>
  html [OPTIONS] -f <input-file> [-o <output-file>]
  html [OPTIONS] (pipe input)

Options:
  -e, --encode     Encode HTML special characters
  -d, --decode     Decode HTML entities
  -s, --smart      Smart detection (auto-encode/decode based on content)
  --full           Use full numeric encoding/decoding (&#NNN; format)
  -f, --file       Process input file
  -o, --output     Write output to file
  --force          Force overwrite without confirmation
  -h, --help       Show this help

Notes:
  - Without -e or -d, defaults to decode (for backward compatibility)
  - With -s, automatically detects and processes appropriately
  - File size limited to 64KB
  - Supports basic (&lt;, &gt;, etc.) and full (&#123;) encoding
  `)
}