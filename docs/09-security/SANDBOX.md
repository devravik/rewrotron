# SANDBOX.md — Containerization & Process Isolation

> **Version:** 1.0  
> **Status:** Security Reference Guide  

---

## 1. Supported Sandbox Modes

Rewrotron supports two execution modes configured in `project.yaml`:

1. **Local Process Sandbox (`mode: local`):**
   - Spawns local PHP-FPM or built-in web servers on randomized localhost loopback ports (`127.0.0.1:40000+`).
   - Traps external network calls via custom `http_proxy` / `CURLOPT_PROXY` environment variables.
2. **Docker Container Sandbox (`mode: docker` - Recommended):**
   - Spawns isolated Docker containers for Source Application, Target Application, and MySQL database.
   - Attaches to a closed bridge network with no external internet gateway (`--internal`).
   - Rewrotron injects mock side-effect receivers within the closed network.
