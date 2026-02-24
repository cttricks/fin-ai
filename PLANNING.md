# Vision

fin is a lightweight, local-first AI utility that enhances search precision.

# Architecture

Browser -> Local HTTP Server -> AI Provider -> Redirect

# Core Modules

1. CLI
2. Config Manager
3. AI Provider Layer
4. HTTP Server
5. Redirect Engine
6. Cache (optional V2)

# Non Goals

- No SaaS backend
- No user tracking
- No analytics
- No cloud storage

# Constraints

- Must start under 100ms (excluding AI call)
- AI call timeout: 10 seconds max
- Fail gracefully to raw search if AI fails

# Current Status (2026-02-24)

- CLI, config storage, AI provider interface, HTTP server, and redirect logic are in place.
- OpenAI and Gemini providers now call real APIs with strict JSON output parsing.

# Future Ideas

- Smart routing (site:github, youtube etc.)
- Query caching
- Search history
- Multi-provider fallback
