# Veil-CLI

Veil is an open-source program written in **Golang** that efficiently scrapes, processes, and manages class and college enrollment data at **De Anza and Foothill College**. It provides seamless class searching, exporting, and automated enrollment monitoring.

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
✅ **Class Monitoring** – Watch class enrollment, get notified of open spots, and auto-enroll in waitlisted courses.  

---

## Prerequisites

Ensure you have **Go (>=1.22.0)** installed. Download it [here](https://go.dev/doc/install).

---

## Installation & Setup

### 1️⃣ Clone the Repository
First, clone the Veil repository to your local machine:

```sh
git clone https://github.com/aandrewduong/veil-cli.git
cd veil-cli
```

### 2️⃣ Install Dependencies
Run the following command to install all required dependencies:

```sh
go mod tidy
```

### 3️⃣ Configure `settings.csv`
1. Copy the example file: `cp settings.csv.example settings.csv`
2. Edit the `settings.csv` file to match your preferences (see **[Configuration](#configuration)** for details).
3. **⚠️ IMPORTANT:** Never commit `settings.csv` to git - it contains sensitive credentials!

---

## Configuration

To function correctly, Veil requires a properly configured **`settings.csv`** file.

### 🔒 Security: Protecting Your Credentials

**Never commit your `settings.csv` or `.credentials` files to git!** They contain sensitive information.

Veil supports three methods for providing your username and password (in priority order):

#### Method 1: Environment Variables (Highest Priority)
Set these environment variables before running Veil:
```sh
export VEIL_USERNAME="your_student_id"
export VEIL_PASSWORD="your_password"
go run .
```

Or in a single command:
```sh
VEIL_USERNAME="your_student_id" VEIL_PASSWORD="your_password" go run .
```

#### Method 2: Credentials File (Recommended for Convenience) ⭐
Create a `.credentials` file in the project root:
```sh
cp .credentials.example .credentials
```

Then edit `.credentials` and add your credentials:
```
username=your_student_id
password=your_password
```

This file is automatically gitignored, so you only need to set it up once and it will persist across sessions.

#### Method 3: settings.csv (Fallback)
If neither environment variables nor `.credentials` file are used, Veil will read from `settings.csv`. Make sure this file is in your `.gitignore` (it already is by default).

**Note:** Even when using environment variables or `.credentials`, you still need `settings.csv` for other configuration (Term, Subject, Mode, CRNs, Webhook, etc.). Just leave the Username and Password fields empty or use placeholders in `settings.csv`.

### `settings.csv` Parameters

| Parameter             | Description                                      | Example Value                              |
|----------------------|------------------------------------------------|-------------------------------------------|
| `Username`           | Your FHDA student ID (or use `VEIL_USERNAME` env var) | `00000000` or leave empty if using env vars |
| `Password`           | Your FHDA password (or use `VEIL_PASSWORD` env var) | `TestTestPassword123` or leave empty if using env vars |
| `Term`              | The academic term                              | `2025 Winter De Anza`                     |
| `Subject`           | Subject for class search                      | `MATH`                                    |
| `Mode`              | Task type (e.g., `Signup`, `Watch`)            | `Signup`                                  |
| `CRNs`              | Course Reference Numbers                      | `47520,44412,41846`                       |
| `Webhook`           | Discord Webhook for notifications             | `https://discord.com/api/webhooks/[...]`  |
| `SavedRegistrationTime` | Registration time (auto-updated)       | *(Do not edit manually)*                  |

#### Setting Up a Discord Webhook  
Follow this guide: [How to Create a Discord Webhook](https://hookdeck.com/webhooks/platforms/how-to-get-started-with-discord-webhooks).

#### Editing `settings.csv`  
Use a spreadsheet editor like [Ron's Editor](https://www.ronsplace.ca/products/ronseditor) or **Google Sheets** for easy modifications.

---

## Compilation

To compile Veil, run:

```sh
bash build.sh
```

---

## Usage

Run the program using:

```sh
go run .
```

Or, if you've compiled it:

```sh
./veil-cli
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
I want Veil to **automatically enroll** me when my registration opens.  
1. Set `Mode` to **`Signup`** and fill in `settings.csv`.  
2. To fully automate registration, first run **Signup** or **Release** mode to save the registration time.  
3. The program will **sleep** until 5 minutes before your registration time, then attempt to enroll you.  

---

### 📌 Scenario 2: Monitoring a Waitlisted Class  
I want to enroll in a class but the **waitlist is full**!  
1. Set `Mode` to **`Watch`** in `settings.csv`.  
2. Run Veil – it will continuously check for openings.  
3. Once a **waitlist spot** is available, **Watch mode** will initiate a **Signup** task to enroll you.  

---

### 📌 Scenario 3: Accessing an Unpublished Course Catalog  
I need the class catalog for a **future term** that isn’t published yet.  
- Simply run **Search mode** – Veil will generate the term ID locally without relying on FHDA's API.  

---

## Screenshots

![image](https://github.com/aandrewduong/veil-v2/assets/135930507/e6e862df-2fde-4015-9095-d9e4818047f3)

---

### 🚀 Contributions & Feedback  
Veil is open-source, and contributions are welcome! Feel free to submit issues, suggestions, or pull requests.
