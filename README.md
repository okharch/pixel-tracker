# Pixel Benchmarking Project

## Purpose

This project benchmarks a high-throughput event tracking backend using Docker Compose. It simulates real-world traffic by sending millions of tracking events to a Go-based HTTP service, backed by PostgreSQL. The goal is to measure how different batch sizes affect ingestion performance.

---

## Project Components

- **pixel_db**: A PostgreSQL 17 container initialized with the required schema.
- **pixel_backend**: A Go-based HTTP API (`/track`) that ingests user events into the database.
- **pixel_benchmark**: A benchmark driver that sends 3 million events using varying batch sizes.

---

## How to Run

To build and run the benchmark:

```bash
docker compose up --force-recreate --build
```

The benchmark runs automatically and shuts down when completed. To manually stop:

```bash
docker compose down
```

If you encounter container name conflicts, run:

```bash
docker container prune
```

---

## Expected Benchmark Output

The benchmark runs 4 configurations with batch sizes of 10, 100, 200, and 400. Each configuration sends 3,000,000 events (from 1,000,000 unique users), split into batches and dispatched by 100 concurrent workers.

| Batch Size | Total Requests | Duration        |
|------------|----------------|-----------------|
| 10         | 300,000        | ~2.44 seconds   |
| 100        | 30,000         | ~1.08 seconds   |
| 200        | 15,000         | ~0.92 seconds   |
| 400        | 7,500          | ~0.84 seconds   |

Performance improves significantly with larger batch sizes due to fewer HTTP and DB writes.

---

## Notes

⚠️ `POSTGRES_HOST_AUTH_METHOD=trust` is used for simplicity but is **not recommended for production**. Set a password instead if security is needed.

📦 The benchmark resets the `track_events` table before each batch run.

