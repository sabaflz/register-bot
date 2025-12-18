# Register Bot

Register Bot is an open-source program written in **Golang** that efficiently scrapes, processes, and manages class and college enrollment data at **De Anza and Foothill College**. It provides seamless class searching, exporting, and automated enrollment monitoring.

Updated for new MyPortal changes.

[![LinkedIn](https://img.shields.io/badge/LinkedIn-Andrew%20Duong-blue)](https://www.linkedin.com/in/andrew-duong-3a9931259/)

---

## Table of Contents

- [Key Features](#key-features)
- [Prerequisites](#prerequisites)
- [Installation & Setup](#installation--setup)
- [Configuration](#configuration)
- [Compilation](#compilation)
- [Usage](#usage)
- [Modes](#modes)
- [Example Scenarios](#example-scenarios)

---

## Key Features

✅ **Class Search & Export** – Search for classes and export results in CSV format.  
✅ **Unofficial Transcript** – Retrieve and export previously enrolled courses.  
✅ **Automated Enrollment** – Enroll in classes at lightning speed.  
✅ **Class Monitoring & Auto-Enrollment** – Watch class enrollment, get notified of open spots, and auto-enroll immediately.  
✅ **Drop & Add (Swapping)** – Automatically drop one course while adding another in a single transaction.
✅ **Multi-College Support** – Run tasks for De Anza and Foothill simultaneously.

---

## Prerequisites

Ensure you have **Go (>=1.22.0)** installed. Download it [here](https://go.dev/doc/install).

---

## Installation & Setup

### 1️⃣ Clone the Repository
First, clone the Register Bot repository to your local machine:

```sh
git clone https://github.com/sabaflz/register-bot.git
cd register-bot
```

### 2️⃣ Install Dependencies
Run the following command to install all required dependencies:

```sh
go mod tidy
```

### 3️⃣ Configure `settings.csv`
1. Copy the example file: `cp config/examples/settings.csv.example config/settings.csv`
2. Edit the `config/settings.csv` file to match your preferences (see **[Configuration](#configuration)** for details).
3. **⚠️ IMPORTANT:** Never commit `config/settings.csv` to git - it contains sensitive credentials!

---

## Configuration

To function correctly, Register Bot requires a properly configured **`config/settings.csv`** file.

### 🔒 Security: Protecting Your Credentials

**Never commit your `config/settings.csv` or `config/.credentials` files to git!** They contain sensitive information.

Register Bot supports three methods for providing your username and password (in priority order):

#### Method 1: Environment Variables (Highest Priority)
Set these environment variables before running Register Bot:
```sh
export REGISTER_BOT_USERNAME="your_student_id"
export REGISTER_BOT_PASSWORD="your_password"
go run .
```

Or in a single command:
```sh
REGISTER_BOT_USERNAME="your_student_id" REGISTER_BOT_PASSWORD="your_password" go run .
```

#### Method 2: Credentials File (Recommended for Convenience) ⭐
Create a `config/.credentials` file:
```sh
cp config/examples/.credentials.example config/.credentials
```

Then edit `config/.credentials` and add your credentials:
```
username=your_student_id
password=your_password
webhook=https://discord.com/api/webhooks/YOUR_WEBHOOK_URL_HERE
```

This file is automatically gitignored, so you only need to set it up once and it will persist across sessions.

#### Method 3: settings.csv (Fallback)
If neither environment variables nor `config/.credentials` file are used, Register Bot will read from `config/settings.csv`. Make sure this file is in your `.gitignore` (it already is by default).

**Note:** Even when using environment variables or `config/.credentials`, you still need `config/settings.csv` for other configuration (Term, Subject, Mode, CRNs, SavedRegistrationTime). Username, Password, and Webhook are now stored in `config/.credentials` for security.

### `settings.csv` Parameters

| Parameter             | Description                                      | Example Value                              |
|----------------------|------------------------------------------------|-------------------------------------------|
| `Term`              | The academic term                              | `2025 Winter De Anza`                     |
| `Subject`           | Subject for class search                      | `MATH`                                    |
| `Mode`              | Task type (e.g., `Signup`, `Watch`)            | `Signup`                                  |
| `CRNs`              | Course Reference Numbers                      | `47520,44412,41846`                       |
| `SavedRegistrationTime` | Registration time (auto-updated)       | *(Do not edit manually)*                  |
| `DropCRNs`          | CRNs to drop before registering (optional)    | `32425`                                   |

**Note:** Username, Password, and Webhook are now stored in `config/.credentials` file (see [Security section](#-security-protecting-your-credentials) above).

#### Setting Up a Discord Webhook  
Follow this guide: [How to Create a Discord Webhook](https://hookdeck.com/webhooks/platforms/how-to-get-started-with-discord-webhooks).

#### Editing `config/settings.csv`  
Use a spreadsheet editor like [Ron's Editor](https://www.ronsplace.ca/products/ronseditor) or **Google Sheets** for easy modifications.

---

## Compilation

To compile Register Bot, run:

```sh
bash scripts/build.sh
```

---

## Usage

Run the program using:

```sh
go run .
```

Or, if you've compiled it:

```sh
./bin/register-bot
```

---

## Modes

| Mode      | Description |
|-----------|------------|
| **Release**  | Similar to `Signup` mode, but waits until **(SavedRegistrationTime - 5 minutes)** before execution (e.g., runs at 7:55 AM if your registration opens at 8:00 AM). Useful for overnight automation. |
| **Signup**   | Enrolls in courses using specified **CRNs**. |
| **Search**   | Searches all available sections for a given term and subject. |
| **Transcript** | Exports your unofficial transcript (previously enrolled courses). |
| **Watch**    | Monitors enrollment availability, notifies you when a spot opens, and attempts to enroll you in the waitlist automatically. |

---

## Example Scenarios

### 📌 Scenario 1: Auto-Enrollment on Registration Day  
I want Register Bot to **automatically enroll** me when my registration opens.  
1. Set `Mode` to **`Signup`** and fill in `config/settings.csv`.  
2. To fully automate registration, first run **Signup** or **Release** mode to save the registration time.  
3. The program will **sleep** until 5 minutes before your registration time, then attempt to enroll you.  

---

### 📌 Scenario 2: Monitoring a Waitlisted Class  
I want to enroll in a class but the **waitlist is full**!  
1. Set `Mode` to **`Watch`** in `config/settings.csv`.  
2. Run Register Bot – it will continuously check for openings.  
3. Once a **waitlist spot** is available, **Watch mode** will initiate a **Signup** task to enroll you.  

---

### 📌 Scenario 3: Accessing an Unpublished Course Catalog  
I need the class catalog for a **future term** that isn't published yet.  
- Simply run **Search mode** – Register Bot will generate the term ID locally without relying on FHDA's API.  

---

## Screenshots

![image](https://github.com/aandrewduong/veil-v2/assets/135930507/e6e862df-2fde-4015-9095-d9e4818047f3)

---

### 🚀 Contributions & Feedback  
Register Bot is open-source, and contributions are welcome! Feel free to submit issues, suggestions, or pull requests.
