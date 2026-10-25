# Go Backend Utilities (Miscellaneous)

> **Production-Ready Go Algorithms, Security Helpers & Backend Patterns**

[![Go Version](https://img.shields.io/badge/Go-Standard_Library-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Security](https://img.shields.io/badge/Security-Crypto_%26_JWT-red?style=flat)](https://golang.org/pkg/crypto/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A curated repository of practical, reusable Go modules and utilities extracted from real-world backend architectures. Includes authentication workflows, cryptography helpers, scheduling routines, and data processing scripts.

---

## 📦 What's Inside

| Module | Description |
| :--- | :--- |
| `signIn.go` / `login.go` | Secure authentication flows with credential verification |
| `schemas.go` | Structured data models and JSON payload validation |
| `middleware.go` | HTTP middleware for route authorization and request auditing |
| `email.go` | SMTP integration for password recovery and notifications |
| `qr.go` | Two-factor authentication (2FA) QR code image generator |
| `generate_pins.go` | Cryptographically random PIN code generation |
| `verify_card.go` | Luhn-based payment card validation logic |
| `backup_scheduler.go` | Automated scheduled database backup routine |
| `file_operations.go` | Streamlined CSV parsing, data extraction, and export |

---

## 🚀 Usage Example

```go
package main

import (
    "fmt"
)

func main() {
    // Example: generate cryptographically secure verification PIN
    pin := GeneratePin(6)
    fmt.Printf("Generated OTP PIN: %s\n", pin)
}
```

---

## 📄 License
MIT License.
