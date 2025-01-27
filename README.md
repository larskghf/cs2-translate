# 🎮 CS2 Chat Translator

[![Build and Push Release](https://github.com/larskghf/cs2-translate/actions/workflows/release.yml/badge.svg)](https://github.com/larskghf/cs2-translate/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/larskghf/cs2-translate)](https://github.com/larskghf/cs2-translate/releases/latest)
[![License](https://img.shields.io/github/license/larskghf/cs2-translate)](https://github.com/larskghf/cs2-translate/blob/main/LICENSE)

A simple Go application that monitors Counter-Strike 2's console output and translates chat messages in real-time using the DeepL API.

## ✨ Features

- 🔄 Real-time monitoring of CS2 chat messages
- 🌍 Automatic translation using DeepL API
- 💬 Supports all CS2 chat types (All Chat, Team Chat)
- 🌐 Language-independent chat detection


## 📋 Prerequisites

- 🔑 A DeepL API key ([Get one here](https://www.deepl.com/pro-api))
- 🎲 Counter-Strike 2 with `-condebug` launch option enabled

## 🚀 Installation

### Option 1: Download pre-built binary (recommended)

1. Download the latest release for your operating system from the [Releases](https://github.com/larskghf/cs2-translate/releases) page
2. Create a `config.json` file in the same directory as the executable (see Configuration section)

### Option 2: Build from source

Requirements:
- 🔧 Go 1.21 or higher

Steps:
1. Clone the repository:
```bash
git clone https://github.com/larskghf/cs2-translate.git
cd cs2-translate
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build
```

4. Copy `config.example.json` to `config.json` and edit it with your settings:
```bash
cp config.example.json config.json
```

5. Edit `config.json` with your settings:
```json
{
    "deepl_api_key": "your-api-key-here",
    "target_language": "DE",
    "console_log_path": "C:\\Program Files (x86)\\Steam\\steamapps\\common\\Counter-Strike Global Offensive\\game\\csgo\\console.log",
    "debug": false,
    "own_name": "YourPlayerName"
}
```
⚠️ Note: For Windows paths in the config file, you need to use double backslashes `\\` as shown above.

## 🎯 Usage

1. Enable console logging in CS2 by adding `-condebug` to your launch options in Steam

2. Run the application:
```bash
# Windows
cs2-translate.exe  # Press Enter to exit when done

# Linux
./cs2-translate
```

⚠️ **Windows Security Note**: Since this is an unsigned executable, Windows Defender might show a warning. You can safely run the application by:
1. Click "More info" in the warning dialog
2. Select "Run anyway"
3. Optionally, add the folder to Windows Defender exclusions

The application will now monitor your CS2 chat and translate messages in real-time.

Example output:
```
[ALL] Player1: Привет всем!
[T] Player2: need help on A!
```

## ⚙️ Configuration

- 🔑 `deepl_api_key`: Your DeepL API key
- 🌍 `target_language`: Target language code (e.g., "DE" for German, "EN-US" for English, etc.)
- 📁 `console_log_path`: Path to your CS2 console log file
- ⚠️ `debug`: Set to true to see detailed debug information
- 👤 `own_name`: Your in-game player name (messages from this player won't be translated)

Default console.log paths:
```
Windows: C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\console.log
Linux:   ~/.steam/steam/steamapps/common/Counter-Strike Global Offensive/game/csgo/console.log
```
⚠️ Remember: When using these paths in config.json, Windows paths need double backslashes (\\).

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Commit Message Format

This repository follows the [Conventional Commits](https://www.conventionalcommits.org/) specification. Commit messages must follow this format:

```
type(scope): description

Examples:
feat(api): add new endpoint
fix: resolve memory leak
docs: update README
```

Requirements:
- Type must be one of:
  - `feat`: New features
  - `fix`: Bug fixes
  - `docs`: Documentation changes
  - `style`: Code style changes (formatting, etc)
  - `refactor`: Code refactoring
  - `test`: Adding or updating tests
  - `chore`: Maintenance tasks
- Scope is optional but must be lowercase if provided
- Description must be at least 10 characters long
- The commit hook will automatically verify these requirements

The Git hooks are automatically set up through the repository's Git attributes. If you're having issues with the commit hook not being executable, you can manually set it:

```bash
git config core.hooksPath .githooks
chmod +x .githooks/commit-msg
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 