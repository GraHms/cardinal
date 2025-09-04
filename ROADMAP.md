# Cardinal Roadmap 🧭

This document outlines the near-term and long-term direction of **Cardinal**, the minimal USSD framework for Go.
Our guiding principles are: **minimalism**, **developer ergonomics**, and **production readiness**.

---

## 🎯 Guiding Principles

1. **Keep the Core Austere**
   Cardinal should remain small: `Request`, `Reply`, `Session`, `Engine`.
   Everything else (stores, transports, middlewares) must be pluggable.

2. **USSD-Native Semantics**
   Use terms natural to USSD developers: `SHOW`, `INPUT`, `Redirect`, `Menu`.
   Avoid leaking HTTP jargon or aggregator quirks into the developer API.

3. **Predictability & Testability**
   Deterministic flows, consistent menu conventions, and built-in simulation for BDD tests.

4. **Practical Production Needs**
   Timeouts, retries, idempotency, observability — built with real telco environments in mind.

---

## 📌 Short-Term (v0.2.x – v0.3.x)

* **Redis Store**
  Production-grade session persistence with TTL, cluster support, and Lua-based atomic ops.

* **Middleware Chain**
  HTTP-like `Use(...)` to attach logging, rate-limit, and metrics around flows.

* **Pagination Helper**
  A tiny abstraction for “8) Prev / 9) Next” in long menus.

* **Improved Testkit**

    * DSL for scripted flows: `Flow("Buy Airtime").Step("home").Send("2").Expect("Enter amount")`.
    * Snapshot diffing for expected screens.

* **Examples Expansion**

    * Banking mini-flow (balance, mini-statement, transfer).
    * Airtime/data bundle catalog with confirmation.

---

## 🚀 Mid-Term (v0.4.x – v0.6.x)

* **CLI Tooling (`cardinal`)**

    * `cardinal new app` → scaffold project layout.
    * `cardinal gen menu` → generate boilerplate for menu flows.

* **Documentation Site**

    * Tutorials: *basic app*, *param routes*, *session persistence*, *testing*.
    * Visual flow diagrams (drawn from route definitions).

* **Aggregator Adapters**
  Out-of-the-box handlers for popular USSD aggregators (Africa’s Talking, Infobip, MTN, Vodacom), normalizing vendor quirks.

* **Observability**

    * Structured logs `{sid, msisdn, path, latency_ms}`.
    * Optional Prometheus metrics (requests, active sessions, errors).

* **i18n / Multi-language Support**
  Light abstraction to translate menu strings.

---

## 🌍 Long-Term (v1.x)

* **Pluggable Encoders**
  GSM-7, UCS-2, and auto-splitting for multi-part messages.

* **Form Helper**
  Multi-field capture within a flow, with built-in validation and retries.

* **Enterprise Hardening**

    * Idempotent side-effects with outbox pattern.
    * Graceful handling of duplicate aggregator retries.
    * Security features: IP whitelisting, optional HMAC signatures.

* **Flow Introspection API**
  Ability to query routes and menu structures at runtime (for documentation and monitoring).

* **Community Ecosystem**
  Encourage external stores (Mongo, Postgres), middlewares (auth, A/B testing), and examples.

---

## 🛣 Release Cadence

* **Patch releases** (`v0.x.y`) every \~2 weeks (bug fixes, minor enhancements).
* **Minor releases** (`v0.y.0`) every \~2–3 months (new modules, features).
* **v1.0.0** once Redis, middleware, pagination, and docs are stable and tested in production.

---

## 🏛 Governance

Cardinal is **MIT-licensed** and open to contributions.

* **Core decisions**: guided by simplicity and production viability.
* **Pull requests**: must include tests and docs.
* **Discussions**: roadmap proposals tracked in GitHub Issues.

---

## 📜 Closing Note

Cardinal is not just a library; it’s a **developer compass** for USSD.
Its role is to **abstract complexity**, **enforce clarity**, and **ensure testability** — while staying out of your way when you need full control.

