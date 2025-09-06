# Cardinal Roadmap 🧭

This document outlines the near-term and long-term direction of **Cardinal**, the minimal USSD framework for Go.
Our guiding principles are: **minimalism**, **developer ergonomics**, and **production readiness**.

---

## 🎯 Guiding Principles

1. **Keep the Core Austere** ✅
   Cardinal remains small: `Request`, `Reply`, `Session`, `Engine`.
   Everything else (stores, transports, middlewares) is pluggable.

2. **USSD-Native Semantics** ✅
   Terms natural to USSD developers: `SHOW`, `INPUT`, `Redirect`, `Menu`.
   No HTTP jargon or aggregator leaks.

3. **Predictability & Testability** ✅
   Deterministic flows, consistent menu conventions, built-in testkit for BDD.

4. **Practical Production Needs** 🔄
   Timeouts, retries, idempotency, observability — partly delivered, expanding.

---

## 📌 Short-Term (v0.2.x – v0.3.x)

* **Redis Store**
  Production-grade session persistence with TTL, cluster support, and Lua-based atomic ops.
  ⏳ *Planned*

* **Middleware Chain** ✅
  HTTP-like `Use(...)` for logging, recovery, rate-limit, and metrics.

* **Per-Route Middleware** ✅
  `SHOWWith(...)` and `INPUTWith(...)` for route-specific concerns.

* **Router Groups** ✅
  Prefix-based grouping with shared middlewares, supporting nesting.

* **Pagination Helper**
  Abstraction for “8) Prev / 9) Next” in long menus.
  ⏳ *Planned*

* **Improved Testkit**

    * DSL for scripted flows
    * Snapshot diffing
      ⏳ *Planned*

* **Examples Expansion**

    * Banking mini-flow
    * Airtime/data bundle catalog
    * **Cardinal Wallet (balances, transfers, history)** ✅

---

## 🚀 Mid-Term (v0.4.x – v0.6.x)

* **CLI Tooling (`cardinal`)**

    * `cardinal new app`
    * `cardinal gen menu`
      ⏳ *Planned*

* **Documentation Site**
  Tutorials and visual flow diagrams.
  ⏳ *Planned*

* **Aggregator Adapters**
  Built-ins for Africa’s Talking, Infobip, MTN, Vodacom.
  ⏳ *Planned*

* **Observability**

    * Structured logs `{sid, msisdn, path, latency_ms}`
    * Prometheus metrics middleware
      ⏳ *Planned*

* **i18n / Multi-language Support**
  ⏳ *Planned*

---

## 🌍 Long-Term (v1.x)

* **Pluggable Encoders** (GSM-7, UCS-2, multipart) ⏳
* **Form Helper** (multi-field capture, validation, retries) ⏳
* **Enterprise Hardening** (idempotent side-effects, retry safety, HMAC) ⏳
* **Flow Introspection API** ⏳
* **Community Ecosystem** (external stores, middlewares, examples) 🔄 already emerging with Wallet and emulator.

---

## 🛣 Release Cadence

* **Patch releases** (`v0.x.y`) every \~2 weeks.
* **Minor releases** (`v0.y.0`) every \~2–3 months.
* **v1.0.0** once Redis, pagination, observability, and docs are hardened in production.

---

## 🏛 Governance

Cardinal is **MIT-licensed** and open to contributions.

* Core decisions guided by simplicity and production viability.
* Pull requests must include tests and docs.
* Roadmap proposals tracked in GitHub Issues.

---

## 📜 Closing Note

Cardinal is not just a library; it’s a **developer compass** for USSD.
It abstracts complexity, enforces clarity, and ensures testability — while staying out of your way when you need full control.

