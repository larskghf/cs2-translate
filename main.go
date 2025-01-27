package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	DeepLAPIKey    string `json:"deepl_api_key"`
	TargetLanguage string `json:"target_language"`
	ConsoleLogPath string `json:"console_log_path"`
	Debug          bool   `json:"debug"`
	OwnName        string `json:"own_name"`
}

type DeepLResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

func loadConfig() (*Config, error) {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func translateText(text, apiKey, targetLang string, debug bool) (string, error) {
	if debug {
		fmt.Printf("Debug: Attempting to translate text: %q\n", text)
		fmt.Printf("Debug: Using API Key: %s...\n", apiKey[:10])
		fmt.Printf("Debug: Target language: %s\n", targetLang)
	}

	apiURL := "https://api-free.deepl.com/v2/translate"
	data := url.Values{}
	data.Set("text", text)
	data.Set("target_lang", targetLang)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		if debug {
			fmt.Printf("Debug: Error creating request: %v\n", err)
		}
		return "", err
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if debug {
			fmt.Printf("Debug: Error making request: %v\n", err)
		}
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if debug {
			fmt.Printf("Debug: Error reading response: %v\n", err)
		}
		return "", err
	}

	if debug {
		fmt.Printf("Debug: DeepL API response: %s\n", string(body))
	}

	var deepLResp DeepLResponse
	if err := json.Unmarshal(body, &deepLResp); err != nil {
		if debug {
			fmt.Printf("Debug: Error parsing response: %v\n", err)
		}
		return "", err
	}

	if len(deepLResp.Translations) == 0 {
		if debug {
			fmt.Printf("Debug: No translations received\n")
		}
		return "", fmt.Errorf("no translation received")
	}

	if debug {
		fmt.Printf("Debug: Translation successful: %q\n", deepLResp.Translations[0].Text)
	}
	return deepLResp.Translations[0].Text, nil
}

func extractChatMessage(line string) (string, string, string, bool) {
	// Clean the line from any trailing whitespace and newlines
	line = strings.TrimSpace(line)

	// Match any line that has exactly two spaces between timestamp and square brackets
	// Format: "01/27 20:28:10  [ANY_TEXT] playername [TOT]: message"
	chatRegex := regexp.MustCompile(`^\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}  \[([^\]]+)\]\s+([^:]+?)(?:\s+\[TOT\])?:\s*(.+)$`)

	matches := chatRegex.FindStringSubmatch(line)
	if matches == nil {
		return "", "", "", false
	}

	chatType := matches[1] // Use original chat type from the message
	player := strings.TrimSpace(matches[2])
	message := strings.TrimSpace(matches[3])
	return chatType, player, message, true
}

func analyzeLogFile(path string, debug bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	if debug {
		fmt.Printf("Debug: Analyzing current log file content...\n")
	}
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(line, "]") && strings.Contains(line, ":") {
			if debug {
				fmt.Printf("Debug: Potential chat line %d: %q\n", lineNum, line)
			}
			if chatType, player, message, ok := extractChatMessage(line); ok && debug {
				fmt.Printf("Debug: Valid chat message found - Type: %q, Player: %q, Message: %q\n", chatType, player, message)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	return nil
}

func monitorFile(config *Config) error {
	fmt.Printf("🎮 CS2 Chat Translator (Target: %s)\n", config.TargetLanguage)
	fmt.Printf("📝 Watching: %s\n\n", config.ConsoleLogPath)

	// Analyze current log content
	if err := analyzeLogFile(config.ConsoleLogPath, config.Debug); err != nil {
		log.Printf("Error analyzing log file: %v", err)
	}

	// Initialize lastSize with current file size
	file, err := os.Open(config.ConsoleLogPath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("error getting file stats: %v", err)
	}

	lastSize := stat.Size()
	file.Close()

	if config.Debug {
		fmt.Printf("Debug: Initial file size: %d bytes\n", lastSize)
		fmt.Printf("Debug: Starting file watch loop...\n")
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		file, err := os.Open(config.ConsoleLogPath)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			continue
		}

		stat, err := file.Stat()
		if err != nil {
			log.Printf("Error getting file stats: %v", err)
			file.Close()
			continue
		}

		currentSize := stat.Size()
		if currentSize != lastSize {
			if config.Debug {
				fmt.Printf("Debug: File size changed: %d bytes (previous: %d bytes)\n", currentSize, lastSize)
			}
			processFileContent(file, &lastSize, currentSize, config)
		}
		file.Close()
	}

	return nil
}

func cleanPlayerName(name string) string {
	// Remove Left-to-Right Mark
	name = strings.ReplaceAll(name, "\u200e", "")
	
	// Handle all variants of @ symbols and split at first occurrence
	atSigns := []string{"@", "﹫", "@", "＠", "＠"}
	for _, at := range atSigns {
		if idx := strings.Index(name, at); idx >= 0 {
			name = name[:idx]
			break
		}
	}
	
	// Replace multiple spaces with single space and trim
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	return strings.TrimSpace(name)
}

func processFileContent(file *os.File, lastSize *int64, currentSize int64, config *Config) {
	if config.Debug {
		fmt.Printf("Debug: Current file size: %d bytes (previous: %d bytes)\n", currentSize, *lastSize)
	}

	// If file size is smaller than last size, file was truncated
	if currentSize < *lastSize {
		if config.Debug {
			fmt.Printf("Debug: File was truncated, resetting position\n")
		}
		*lastSize = 0
		return
	}

	// Only process if we have new content
	if currentSize > *lastSize {
		if config.Debug {
			fmt.Printf("Debug: New content detected, processing...\n")
		}
		
		// Seek to last read position
		_, err := file.Seek(*lastSize, 0)
		if err != nil {
			log.Printf("Error seeking in file: %v", err)
			return
		}

		// Read new content
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if config.Debug {
				fmt.Printf("Debug: Processing line: %q\n", line)
			}
			if chatType, player, message, ok := extractChatMessage(line); ok {
				if config.Debug {
					fmt.Printf("Debug: Chat message detected - Type: %q, Player: %q, Message: %q\n", chatType, player, message)
				}
				
				// Clean the player name
				cleanedPlayer := cleanPlayerName(player)

				// Skip translation if it's our own message
				if config.OwnName != "" && cleanPlayerName(config.OwnName) == cleanedPlayer {
					if config.Debug {
						fmt.Printf("Debug: Skipping own message from %q\n", cleanedPlayer)
					}
					continue
				}
				
				// Translate the message
				translatedText, err := translateText(message, config.DeepLAPIKey, config.TargetLanguage, config.Debug)
				if err != nil {
					log.Printf("Error translating message: %v", err)
					continue
				}

				// Print the translated message with cleaned player name
				fmt.Printf("[%s] %s: %s\n", chatType, cleanedPlayer, translatedText)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading file: %v", err)
			return
		}

		*lastSize = currentSize
	} else if config.Debug {
		fmt.Printf("Debug: No new content to process\n")
	}
}

func waitForEnter() {
	if runtime.GOOS == "windows" {
		fmt.Println("\nPress Enter to exit...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
}

func main() {
	if runtime.GOOS == "windows" {
		// Redirect log output to stdout for Windows
		log.SetOutput(os.Stdout)
	}

	config, err := loadConfig()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		waitForEnter()
		os.Exit(1)
	}

	if err := monitorFile(config); err != nil {
		log.Printf("Error monitoring file: %v", err)
		waitForEnter()
		os.Exit(1)
	}

	// Always wait for Enter on Windows before exiting
	waitForEnter()
}

