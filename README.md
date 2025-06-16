# Telegram Course Enrollment Bot 🎓

A Telegram bot that provides real-time insights about class enrollments, including current enrollment numbers, available seats, and section details.

## Features

- 🔍 Real-time course enrollment status
- 📌 Course subscription system
- 👀 Section-wise enrollment details
- 🔄 Auto-updates every 10 minutes
- 💾 Persistent storage with SQLite

## Commands

- `/start` - Get started with the Course Bot
- `/subscribe COURSE` - Subscribe to a course (e.g. `/subscribe CSCI 151`)
- `/unsubscribe COURSE` - Unsubscribe from a course
- `/list` - View your subscribed courses with interactive buttons
- `/showall` - View detailed info for all subscribed courses

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Create `.env` file:
```env
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_ADMIN_ID=your_admin_id
COURSES_API_URL=your_api_url
```

3. Run the bot:
```bash
go run cmd/api/main.go
```

## Configuration

The bot uses environment variables for configuration:
- `TELEGRAM_BOT_TOKEN`: Your Telegram bot token from BotFather
- `TELEGRAM_ADMIN_ID`: Your Telegram user ID
- `COURSES_API_URL`: API endpoint for course data