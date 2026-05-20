
---

# Broccoli Bank API - Core Backend

## Overview

Broccoli Bank is a high-performance, concurrency-safe core banking API built in pure Go. This project demonstrates enterprise-grade financial architecture by implementing strict ACID transactions, a double-entry ledger, and robust deadlock prevention. It was built to solve the hardest parts of backend engineering: state management, concurrent data access, and financial security.

## Tech Stack

* **Language**: Go (Standard Library `net/http` for routing and API handlers).
* **Database**: MySQL (Community Edition).
* **Database Access**: `database/sql` with `go-sql-driver/mysql`. *Note: Intentionally avoids ORMs to maintain absolute control over SQL query execution and lock management*.
* **Security**: `golang.org/x/crypto/bcrypt` (hashing) and `github.com/golang-jwt/jwt/v5` (authentication).

## Core Features & Architectural Decisions

### 1. The Double-Entry Ledger

Financial software requires an append-only audit log. This API utilizes a strict Header/Lines database pattern. Every financial movement (deposit, transfer, withdrawal) creates a transaction header and generates immutable, perfectly opposing debit and credit ledger entries to guarantee a mathematically zero-sum system.

### 2. Concurrency & Deadlock Prevention

Handling user-to-user transfers concurrently introduces significant race conditions and deadlocks. This API solves them by:

* **Pessimistic Locking:** Utilizing `SELECT ... FOR UPDATE` to strictly lock account rows during balance mutations.

* **Lock Ordering:** Mathematically eliminating deadlocks by pre-sorting account IDs and consistently locking the lowest ID first, forcing simultaneous bidirectional transfers into a safe, single-file line.

* **The "Hot Row" Bypass:** External deposits are handled via a webhook (mimicking Plaid/Stripe) that utilizes an "Insert-Only Ledger" pattern for the central bank system account, bypassing row contention and allowing infinite horizontal scaling for deposits.

### 3. Idempotency & Resilience

This API features a custom, dependency-free exponential backoff and jitter retry loop designed to catch and gracefully retry safe, transient database errors (like `ER_LOCK_WAIT_TIMEOUT` and `ER_LOCK_DEADLOCK`).

### 4. Defensive Security

* **Strict Validation:** Enforces 72-byte bcrypt limits to prevent denial-of-service vulnerabilities , normalizes user input, and immediately rejects negative transaction amounts.

* **Stateless Auth:** Secures endpoints via 10-minute short-lived JWTs.

* **Safe Context Injection:** Uses unexported custom types for HTTP context keys to prevent middleware collisions.

## Testing Strategy

The system's integrity is validated by a pragmatic "Test Pyramid" running against a live MySQL test instance (`broccoli_test`), ensuring the database constraints are proven in reality rather than mocked:

1. **Database Lifecycle Tests:** Validates constraints, foreign keys, and transaction rollbacks directly at the storage layer.

1. **API Edge-Case Isolation:** Table-driven tests that hammer boundary conditions (negative amounts, missing tokens, invalid payloads) to guarantee "Fail Fast" principles without redundant setup bloat.

1. **The Zero-Sum E2E Journey:** A holistic end-to-end integration test that orchestrates user creation, webhook fund ingestion, concurrent internal transfers, and withdrawals. It concludes with a final SQL aggregation query to prove the entire ecosystem maintained a perfect zero-sum state.

---
