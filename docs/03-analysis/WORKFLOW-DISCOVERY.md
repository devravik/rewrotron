# WORKFLOW-DISCOVERY.md — Multi-Step Workflow Mining & State Transition Modeling

> **Version:** 1.0  
> **Status:** Analysis Reference Guide  

---

## 1. Defining Multi-Step Workflows

A workflow represents a coherent, multi-request business process. While individual routes can be migrated as standalone operations, multi-step workflows must maintain session continuity, temporary database state, and transaction boundaries across steps.

Examples of critical workflows:
- **Authentication & Recovery:** Login -> MFA Challenge -> Session Establishment -> Logout.
- **E-Commerce Checkout:** Cart Review -> Address Entry -> Payment Processing -> Order Confirmation.
- **Multistage Forms:** Step 1 Personal Info -> Step 2 Documents Upload -> Step 3 Submission.

---

## 2. Mining Heuristics

Rewrotron discovers workflows using three correlation heuristics:

1. **Session State Continuity:** Sequences of HTTP requests that mutate and consume shared session keys (e.g. `step_1_data` set in request A, read in request B).
2. **HTML Form Action Chaining:** Static analysis of rendered Blade/HTML views discovering `<form action="/step2">` within the response of `/step1`.
3. **HTTP 302 Redirect Chaining:** Requests terminating in a 302/303 redirect immediately followed by a GET to the target location.

Discovered workflows are stored as `Workflow` records in `project.db` and assigned higher verification priority than isolated read-only GET requests.
