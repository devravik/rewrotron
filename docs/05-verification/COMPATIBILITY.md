# COMPATIBILITY.md — Compatibility Matrix Across Subsystems

> **Version:** 1.0  
> **Status:** Verification Reference Guide  

---

## 1. Subsystem Compatibility Matrix

| Dimension | Legacy CI3 Baseline | Target Laravel 13 Baseline | Verification Strategy |
| :--- | :--- | :--- | :--- |
| **HTTP Routing** | `application/config/routes.php` | `routes/web.php`, `routes/api.php` | Differential HTTP Replay |
| **Request Inputs** | `$this->input->post()` | Form Requests (`$request->validated()`) | Payload Schema Assertion |
| **Session State** | `CI_Session` (`$_SESSION['ci_session']`) | Laravel Session Store | Session Diff Comparator |
| **Flash Data** | `set_flashdata()` | `with('key', 'value')` | Session Flash Diff |
| **Active Record** | `$this->db->get_where()` | Eloquent Models / Query Builder | Database Mutation Diff |
| **View Rendering** | PHP template files | Blade templates (`.blade.php`) | DOM Semantic Comparator |
| **Validation** | CI3 Form Validation library | Laravel Validator | Error Message Normalization |
| **Security CSRF** | `$config['csrf_protection']` | VerifyCsrfToken Middleware | Token Pattern Masking |
