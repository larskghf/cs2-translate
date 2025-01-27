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

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	DeepLAPIKey    string `json:"deepl_api_key"`
	TargetLanguage string `json:"target_language"`
	ConsoleLogPath string `json:"console_log_path"`
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

func translateText(text, apiKey, targetLang string) (string, error) {
	apiURL := "https://api-free.deepl.com/v2/translate"
	data := url.Values{}
	data.Set("text", text)
	data.Set("target_lang", targetLang)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var deepLResp DeepLResponse
	if err := json.Unmarshal(body, &deepLResp); err != nil {
		return "", err
	}

	if len(deepLResp.Translations) == 0 {
		return "", fmt.Errorf("no translation received")
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

func monitorFile(config *Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Watch for changes
	if err := watcher.Add(config.ConsoleLogPath); err != nil {
		return err
	}

	fmt.Printf("🎮 CS2 Chat Translator (Target: %s)\n", config.TargetLanguage)
	fmt.Printf("📝 Watching: %s\n\n", config.ConsoleLogPath)

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

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				file, err := os.Open(config.ConsoleLogPath)
				if err != nil {
					log.Printf("Error opening file: %v", err)
					continue
				}

				// Get current file size
				stat, err := file.Stat()
				if err != nil {
					log.Printf("Error getting file stats: %v", err)
					file.Close()
					continue
				}

				// If file size is smaller than last size, file was truncated
				if stat.Size() < lastSize {
					lastSize = 0
				}

				// Seek to last processed position
				_, err = file.Seek(lastSize, 0)
				if err != nil {
					log.Printf("Error seeking in file: %v", err)
					file.Close()
					continue
				}

				reader := bufio.NewReader(file)
				for {
					line, err := reader.ReadString('\n')
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Printf("Error reading line: %v", err)
						break
					}

					if chatType, player, message, ok := extractChatMessage(line); ok {
						translated, err := translateText(message, config.DeepLAPIKey, config.TargetLanguage)
						if err != nil {
							log.Printf("Error translating message: %v", err)
							continue
						}
						fmt.Printf("[%s] %s: %s\n", chatType, player, translated)
					}
				}

				// Update last processed position
				lastSize = stat.Size()
				file.Close()
			}
		case err := <-watcher.Errors:
			log.Printf("Watcher error: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
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
