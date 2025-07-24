# Social Content DL

A Go application for downloading social media content, including music from Telegram channels.

## Features

- Download music from Telegram channels using your personal account
- Support for audio files, voice messages, and documents
- Automatic file naming and organization
- Duplicate file detection
- No need to add bots to channels

## Getting Started

### Prerequisites

- Go 1.21 or later
- Telegram account (your personal account)

### Setup Telegram API

1. Go to https://my.telegram.org/apps
2. Log in with your phone number
3. Create a new application (or use existing)
4. Note down your `api_id` and `api_hash`

### Installation

1. Clone the repository:
```bash
git clone https://github.com/ehsundar/social-content-dl.git
cd social-content-dl
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set your Telegram credentials:
```bash
export TELEGRAM_PHONE="+1234567890"  # Your phone number
export TELEGRAM_APP_ID="your_api_id"  # From my.telegram.org/apps
export TELEGRAM_APP_HASH="your_api_hash"  # From my.telegram.org/apps
```

4. Run the application:
```bash
# Download all music from a channel
go run main.go telegram channelname

# Download limited number of files
go run main.go telegram channelname 50
```

**Note:** On first run, you'll receive a verification code on your Telegram account. Enter it when prompted.

## Environment Variables

- `TELEGRAM_PHONE` - Your phone number (required, e.g., +1234567890)
- `TELEGRAM_APP_ID` - Your API ID from my.telegram.org/apps (optional, default: 17349)
- `TELEGRAM_APP_HASH` - Your API hash from my.telegram.org/apps (optional, default: desktop)
- `DOWNLOAD_PATH` - Download directory (default: ./downloads)

## Usage Examples

```bash
# Download all music from @musicchannel
go run main.go telegram musicchannel

# Download only 10 files from @musicchannel
go run main.go telegram musicchannel 10

# Use custom download path
export DOWNLOAD_PATH="/path/to/music"
go run main.go telegram musicchannel
```

## Development

To add new dependencies:
```bash
go get <package-name>
```

## License

This project is licensed under the MIT License.
