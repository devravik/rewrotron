# STATIC-ANALYSIS.md — AST Parsing & Reachability Call Graph Architecture

> **Version:** 1.0  
> **Status:** Architecture Reference Guide  

---

## 1. Parser Strategy & Go Implementation

Static analysis of PHP codebases requires parsing potentially legacy PHP (PHP 5.6 through 8.2) without requiring the user to run legacy PHP runtimes for the analysis phase.

1. **Tree-Sitter / Go-Native PHP Parser:** Primary parser implemented in Go using Tree-Sitter PHP grammar or native Go parser for maximum speed, memory safety, and cross-platform compilation.
2. **Companion Reflection Agent (`tools/php/reflect.php`):** For ambiguous dynamic class definitions or complex inheritance hierarchies, a small sandboxed PHP CLI script is invoked to emit structured JSON reflection metadata.

---

## 2. Call Graph Construction & Transitive Reachability

The call graph is represented as a directed graph $G = (V, E)$:
- **Vertices $V$:** Declared functions, class methods, closures, and file script entrypoints.
- **Edges $E$:** Directed invocation edges $(u, v)$ where method $u$ calls method $v$.

### Handling Dynamic Dispatches
In legacy PHP, dynamic dispatches are common:
```php
$model_name = $this->input->post('type') . '_model';
$this->load->model($model_name);
$this->$model_name->process();
```

When static analysis encounters unresolvable variable variables, dynamic class instantiations, or variable method names:
1. The edge is marked as `DYNAMIC_EDGE_UNRESOLVED`.
2. The target potential symbols are classified as `UNKNOWN`.
3. The Mission Manager creates an `OBSERVE_WORKFLOW` mission to resolve the dynamic dispatch at runtime.
