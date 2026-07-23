package chatsummary

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// ── File layout ─────────────────────────────────────────────────
//
//   {baseDir}/{sessionID}/
//     ├── last_summary.txt   ← plain text, overwritten each time
//     ├── messages.jsonl     ← append-only, one JSON message per line
//     └── .lock              ← (reserved for future use)

func sessionDir(baseDir, sessionID string) string {
	return filepath.Join(baseDir, sessionID)
}

// appendMessage writes msg as a JSON line to messages.jsonl,
// creating the directory and file if needed.
func appendMessage(baseDir, sessionID string, msg RawMessage) error {
	dir := sessionDir(baseDir, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(dir, "messages.jsonl"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}

	return nil
}

// readSummary returns the contents of last_summary.txt.
// Returns empty string if the file does not exist.
func readSummary(baseDir, sessionID string) (string, error) {
	data, err := os.ReadFile(filepath.Join(sessionDir(baseDir, sessionID), "last_summary.txt"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// writeSummary overwrites last_summary.txt with the given text.
func writeSummary(baseDir, sessionID, summary string) error {
	return os.WriteFile(
		filepath.Join(sessionDir(baseDir, sessionID), "last_summary.txt"),
		[]byte(summary), 0644)
}

// readRecentMessages returns the last n messages from messages.jsonl.
// If the file contains fewer than n messages, all are returned.
func readRecentMessages(baseDir, sessionID string, n int) ([]RawMessage, error) {
	all, err := readAllMessages(baseDir, sessionID)
	if err != nil {
		return nil, err
	}

	if len(all) <= n {
		return all, nil
	}
	return all[len(all)-n:], nil
}

// readAllMessages reads every line from messages.jsonl.
// Returns an empty slice if the file does not exist.
func readAllMessages(baseDir, sessionID string) ([]RawMessage, error) {
	f, err := os.Open(filepath.Join(sessionDir(baseDir, sessionID), "messages.jsonl"))
	if err != nil {
		if os.IsNotExist(err) {
			return []RawMessage{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var messages []RawMessage
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var msg RawMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
