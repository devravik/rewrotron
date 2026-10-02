# ROUTE-DISCOVERY.md — Static and Dynamic Route Discovery

> **Version:** 1.0  
> **Status:** Analysis Reference Guide  

---

## 1. Route Discovery Strategy

Legacy web frameworks often employ multiple concurrent routing paradigms. For CodeIgniter 3.x, Rewrotron extracts routes via three cooperative mechanisms:

1. **Configured Static Routes:** Parsed from `application/config/routes.php`, including regular expressions (`$route['products/(:num)'] = 'catalog/detail/$1'`) and wildcards.
2. **Implicit Convention Routes:** Inferred from public controller classes and public action methods (`class Products extends CI_Controller` -> `/products`, `/products/view/{id}`).
3. **Observed Traffic Routes:** Extracted from HTTP access logs, recorded HAR files, or dynamic crawler executions discovering obscure legacy endpoints not declared in static configs.

---

## 2. Resolving Ambiguities & Fallbacks

- **Default Controller:** Identified via `$route['default_controller']` in CI3 config.
- **404 Override Controller:** Identified via `$route['404_override']`.
- **Sub-directory Routing:** Controllers nested inside subdirectories (e.g. `application/controllers/admin/Users.php` -> `/admin/users`).
- **HTTP Method Constraints:** Legacy CI3 lacks native HTTP verb constraints on routes. Rewrotron inspects method bodies for checks like `$this->input->method() === 'post'` or `$this->input->is_ajax_request()` to assign HTTP verb constraints to discovered routes.
