# HTML Encoder/Decoder Tool

A fast and lightweight **HTML Encoding and Decoding CLI Tool** written in Go.

This tool allows you to encode and decode HTML entities, perform smart detection, and process both text and files directly from the command line.

---
## 🔧 Options
| Option         | Description                                   |
| -------------- | --------------------------------------------- |
| `-e, --encode` | Encode HTML special characters                |
| `-d, --decode` | Decode HTML entities                          |
| `-s, --smart`  | Smart detection (auto-encode/decode)          |
| `--full`       | Use full numeric encoding/decoding (`&#NNN;`) |
| `-f, --file`   | Process input file                            |
| `-o, --output` | Write output to file                          |
| `--force`      | Force overwrite without confirmation          |
| `-h, --help`   | Show help message                             |

---

## 🚀 Features

- Encode HTML special characters
- Decode HTML entities
- Smart auto-detection mode
- Full numeric encoding/decoding (`&#NNN;`)
- File input/output support
- Pipe support (stdin)
- Safe overwrite confirmation
- Lightweight and fast (Go binary)
- File size limit: 64KB

---

## 📦 Installation

### 1️⃣ Requirements

- Go 1.18+

Check Go version:

```bash
go version
```

```bash
go build -o html main.go
```

```bash
chmod +x html
sudo cp html /usr/local/bin/
```

```bash
html -h
```

### Create a .html file with the basic html structure
```bash
html main
```
- this wil create a html file.


```

```bash
```
