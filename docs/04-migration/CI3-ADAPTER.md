# CI3-ADAPTER.md — CodeIgniter 3.x Source Extraction Adapter Specification

> **Version:** 1.0  
> **Status:** Adapter Specification  

---

## 1. CodeIgniter 3 Execution Model

The CodeIgniter 3 adapter introspects:
- **Routing:** `application/config/routes.php` and dynamic `URI` class segment routing.
- **Controllers:** Subclasses of `CI_Controller` with action methods.
- **Active Record:** `$this->db` calls (`select`, `join`, `where`, `get`, `insert`, `update`, `delete`).
- **Loader:** `$this->load->model()`, `$this->load->library()`, `$this->load->helper()`, `$this->load->view()`.
- **Sessions & Flash Data:** `$this->session->userdata()` and `$this->session->flashdata()`.
- **Hooks System:** `application/config/hooks.php` hooks used for runtime instrumentation injection.

---

## 2. Injected Runtime Observer Hook

The CI3 adapter injects a non-invasive observer hook into `application/config/hooks.php`:

```php
$hook['pre_controller'][] = array(
    'class'    => 'Rewrotron_Observer',
    'function' => 'record_pre_controller',
    'filename' => 'Rewrotron_Observer.php',
    'filepath' => 'hooks'
);
$hook['post_controller'][] = array(
    'class'    => 'Rewrotron_Observer',
    'function' => 'record_post_controller',
    'filename' => 'Rewrotron_Observer.php',
    'filepath' => 'hooks'
);
```

This captures the exact entry and exit state of every request without modifying application source files.
