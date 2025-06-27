# Telegram Course Enrollment Bot 🎓

A Telegram bot that delivers real-time updates on university course enrollments, including current numbers, available seats, and section details. Users can subscribe to courses and receive notifications when spots become available.

## Features

- 🔍 Live course enrollment tracking
- 📬 Subscribe/unsubscribe to course sections
- 👀 Detailed section-wise enrollment info
- 🔄 Automatic updates every 10 minutes
- 💾 Persistent subscriptions with SQLite

## Commands

- `/start` — Welcome and usage instructions
- `/subscribe COURSE [SECTION ...]` — Subscribe to a course and optional sections (e.g. `/subscribe CSCI151 1L 2CLb`)
- `/unsubscribe COURSE` — Unsubscribe from a course
- `/list` — Show your current subscriptions with interactive buttons

## Getting Started

1. **Install dependencies:**
    ```bash
    go mod download
    ```

2. **Create a `.env` file:**
    ```env
    TELEGRAM_BOT_TOKEN=your_bot_token
    TELEGRAM_ADMIN_IDS=123456789,987654321
    COURSES_API_URL=https://your-university.edu/courses.xls
    ```

3. **Run the bot:**
    ```bash
    go run cmd/api/main.go
    ```

## Configuration

Set these environment variables in your `.env` file:
- `TELEGRAM_BOT_TOKEN`: Telegram bot token from BotFather
- `TELEGRAM_ADMIN_IDS`: Comma-separated list of admin Telegram user IDs
- `COURSES_API_URL`: URL to the XLS file or API endpoint with course data
