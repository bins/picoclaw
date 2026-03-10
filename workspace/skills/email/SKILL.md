---
name: email
description: Read and process emails received via the /webhook/email endpoint. Emails
  are stored as daily JSONL files in {workspace}/emails/. Use this skill to list,
  summarize, search, or respond to incoming emails.
---

# Email Skill

Emails pushed to the gateway's `POST /webhook/email` endpoint are appended as
JSON lines to daily files under the workspace.

## Storage Layout

```
~/.picoclaw/workspace/emails/
├── 2026-03-10.jsonl
├── 2026-03-11.jsonl
└── ...
```

Each line is a JSON object:

```json
{"received_at":"2026-03-10T08:32:01Z","from":"alice@example.com","subject":"Meeting notes","content":"Hi, here are the notes..."}
```

## Common Tasks

### List available days

```bash
ls ~/.picoclaw/workspace/emails/
```

### Read today's emails

```bash
cat ~/.picoclaw/workspace/emails/$(date +%Y-%m-%d).jsonl
```

### Pretty-print with jq

```bash
cat ~/.picoclaw/workspace/emails/$(date +%Y-%m-%d).jsonl | jq .
```

### Show only sender and subject

```bash
cat ~/.picoclaw/workspace/emails/$(date +%Y-%m-%d).jsonl \
  | jq -r '[.received_at, .from, .subject] | @tsv'
```

### Search by sender

```bash
grep '"from":"alice@example.com"' ~/.picoclaw/workspace/emails/*.jsonl | jq .
```

### Search by keyword in subject or content

```bash
grep -i "invoice" ~/.picoclaw/workspace/emails/*.jsonl | jq .
```

### Count emails per day

```bash
for f in ~/.picoclaw/workspace/emails/*.jsonl; do
  echo "$(basename $f .jsonl): $(wc -l < $f) emails"
done
```

## Sending a Test Email

```bash
curl -s -X POST http://localhost:18790/webhook/email \
  -H "Content-Type: application/json" \
  -H "X-Pico-Auth: <your-secret>" \
  -d '{"from":"test@example.com","subject":"Hello","content":"Test body"}'
```

## Notes

- The webhook requires `email_webhook.enabled: true` in `config.json`.
- Authentication uses the `X-Pico-Auth` header matched against `email_webhook.secret`.
- Files are created automatically on first write; missing days simply have no file.
- Each file is append-only; records are never modified or deleted automatically.
