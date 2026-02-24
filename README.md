# fin-ai
Local AI-powered search intent optimizer for any browser.

fin is a local AI-powered search intent optimizer.

It converts natural language search input into optimized Google queries using OpenAI or Gemini.

Example:

Input:
find ai upcoming events in Delhi

AI Output:
AI conferences and tech events in Delhi 2026

Redirect:
https://google.com/search?q=AI+conferences+and+tech+events+in+Delhi+2026

---

## Features

- Local HTTP server
- AI query optimization
- OpenAI & Gemini support
- CLI interface
- Secure API key storage
- Zero telemetry
- Browser keyword integration

---

## Installation

go build -o fin ./cmd/fin

---

## Usage

fin run -p 2026

Then configure browser search engine:

Keyword: find
URL:
http://localhost:2026/?text=%s

---

## CLI

fin -v
fin -h
fin openai|gemini -key "<your-key>"
