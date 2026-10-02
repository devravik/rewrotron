# LARAVEL-ADAPTER.md — Laravel 13.x Target Code Generation Adapter Specification

> **Version:** 1.0  
> **Status:** Adapter Specification  

---

## 1. Generation Targets & Mapping Rules

The Laravel 13 adapter converts abstract `BehaviorContract` specifications into modern Laravel 13 components:

| Legacy CI3 Artifact | Abstract Behavior Entity | Generated Laravel 13 Artifact |
| :--- | :--- | :--- |
| `config/routes.php` | `EntryPoint.HTTP` | `routes/web.php` or `routes/api.php` |
| `CI_Controller::method` | `Operation` | Invokable Controller or Action class |
| `$this->input->post()` | `Inputs.parameters` | Dedicated `app/Http/Requests/*Request.php` |
| `$this->db->*` | `DatabaseEffects` | Eloquent Models (`app/Models/`) & Query Builders |
| `$this->session->*` | `SessionEffects` | Laravel Session (`session(['key' => $val])`) |
| `$this->load->view()` | `Response.Body` | Blade Templates (`resources/views/*.blade.php`) |

---

## 2. Retained Relational Database Strategy

In the v1 release, Rewrotron retains the existing relational schema (e.g. MySQL 8.x):
- Eloquent models map to existing table names and primary keys:
  ```php
  protected $table = 'legacy_orders';
  protected $primaryKey = 'order_id';
  public $timestamps = false; // until modern timestamp migration is enabled
  ```
- This prevents simultaneous data corruption while verifying behavioral compatibility.
