# Project context — full-stack calculator

The brief this project was built against. It records what was decided before and
during implementation; the README describes what was actually built. Kept in the
repository alongside PROMPTS.md so the reasoning behind the code is visible, not
just the result.

Context for AI assistance on this project. Read before generating anything.
If you think a decision here is wrong, say so and wait — do not restructure
around it.

**How to respond:** keep answers short. After any output, add two or three sentences explaining what it does and why it was done that way. No preamble, no summaries of work already shown.

---

## 1. The assignment

Build a full-stack calculator: React + TypeScript frontend consuming a Go REST
API that performs the arithmetic. Required operations are addition,
subtraction, multiplication and division; exponentiation, square root and
percentage are optional. The frontend needs an intuitive UI, input validation,
error handling and basic responsive mobile support. The backend exposes REST
endpoints, validates input, handles edge cases such as division by zero and
invalid data, and returns JSON. Both layers need unit tests covering key
functionality. Documentation must cover setup, API usage and design rationale.
A Dockerfile for full-stack deployment is optional. Deliverables are a Git
repository, a README, tests with a coverage report, and the prompts used.

Budget is roughly 2–4 hours, with correctness, clarity and maintainability
prioritised over extra features. AI tooling is explicitly permitted provided
the prompts are shared.

---

## 2. Scope

In scope: addition, subtraction, multiplication, division, exponentiation,
square root and percentage, plus a sign toggle on the keypad.

Out of scope, deliberately: calculation history, memory functions,
parenthesised expressions, an expression parser (the API takes discrete
operands, not expression strings), authentication, persistence, multiple
backend services, state management or component libraries, and routing. Do
not add these and do not suggest them. Anything deferred gets one line in the
README under what I would add with more time.

"Microservice" here means the backend is a separate, independently runnable
and deployable process communicating with the frontend over HTTP/JSON. One
service, not several.

---

## 3. Architecture

    /backend
      /cmd/api            main.go — wiring and startup only
      /internal/calc      pure arithmetic + domain errors, no HTTP
      /internal/handlers  HTTP handlers, decoding, status mapping, routing
      /internal/httpx     CORS middleware
    /frontend
      /src/components  /src/hooks  /src/types  /src/api
    README.md  CONTEXT.md  PROMPTS.md  docker-compose.yml

The load-bearing boundary is `internal/calc` knowing nothing about HTTP —
arithmetic is pure functions returning `(float64, error)`, which keeps the
core testable without a server. `internal/` is compiler-enforced against
outside imports and `cmd/` separates the binary entrypoint from library code.
These two directories come from common Go layout conventions;
`golang-standards/project-layout` is not an official Go standard, so the
README states that only the parts that apply at this scale were used.

Routing is `net/http` with Go 1.22+ patterns (`mux.HandleFunc("POST /api/calculate", h)`),
which yields 404 and 405 without extra code. Zero third-party backend
dependencies.

Response helpers live in `handlers` rather than `httpx` because they are used
by one package; `httpx` holds only middleware.

---

## 4. API contract and decisions to justify

One endpoint: `POST /api/calculate`, taking `{"operation", "a", "b"}` where
`b` is omitted for unary operations, returning `{"result": n}` on success and
`{"error": {"code", "message"}}` on failure.

The alternative was one route per operation, such as `POST /api/add`. A
single endpoint means request decoding, input validation and error mapping
are written once rather than repeated in every handler, and adding an
operation later means adding one case to a switch and one row to a test
table. The cost is that the operation name travels in the request body
instead of the URL path, which is less conventionally RESTful. Both designs
are defensible; the README states which was chosen and what was given up.

Status codes: 200 success; 400 for malformed or empty JSON, trailing data,
unknown operation, missing or non-numeric operand, division by zero, square
root of a negative, and non-finite results; 404 unknown path; 405 wrong
method; 413 body over 1 MiB; 415 missing or non-JSON `Content-Type`; 500
unexpected faults. Division by zero is 400 because the input is what is wrong,
not the server.

Each of these needs its reasoning written in the README, not just the
outcome. **Floating point:** compute in `float64`, round only at the response
boundary, to significant digits rather than decimal places — fixed decimals
destroy small results and overflow on large ones; note that decimal arithmetic
would be mandatory for money but is overkill here. **Non-finite results:**
`json.Marshal` fails on `Inf` and `NaN`, so guard with `math.IsInf` and
`math.IsNaN` before encoding and return 400; marshal before writing the status
so an encoding failure can never follow a 200. **Percentage:** `50 % 200`
means 50% of 200, equalling 100, modeled as a binary operation rather than the
contextual modifier physical calculators implement, because the API has no
expression context. **Operands:** both are pointers in the request type so an
absent operand is distinguishable from an operand of zero. **Validation on
both layers:** the frontend is UX, the backend is authoritative and never
trusts the client. **CORS:** allowed origin from an environment variable, with
the `OPTIONS` preflight handled explicitly. **`http.MaxBytesReader`** on the
request body, and read/write/header timeouts on the server rather than bare
`ListenAndServe`.

---

## 5. Style

**Go** follows Effective Go and Go Code Review Comments, with the Uber style
guide for anything they leave open. `gofmt` and `go vet` clean before every
commit. Errors return early so the happy path stays unindented; sentinel
errors are named `ErrX` with lowercase, unpunctuated messages; short names in
short scopes; accept interfaces, return structs, never a pointer to an
interface; doc comments on every exported identifier.

**Frontend** follows the current react.dev docs with ESLint and Prettier
clean, function components only, no `any`, semantic HTML with real `<button>`
elements and visible focus states, the request/response contract typed once as
a discriminated union so an invalid operand combination is a compile error,
and the API base URL from `VITE_API_URL`. A keypad calculator has no text
input, so "controlled inputs" does not apply: the display is an `<output>`
with `aria-live`, and the error message carries `role="alert"`. Keyboard
support covers digits, operators, Enter for equals, Escape for clear,
Backspace for delete and `n` for sign toggle. Responsive behaviour is CSS grid
with tap targets of at least 44px, verified manually at 375px and on a
physical device rather than tested automatically — the README says "verified
manually".

---

## 6. Tests

Tests cover branches, not arithmetic — `Add(2,3)` and `Add(-2,-3)` exercise
the same line, so the second earns nothing. Every case should answer what bug
could ship if it were deleted.

**`internal/calc`**, table-driven with `t.Run`: one happy path per operation;
division by zero returning the sentinel error; `0**0` and negative exponents;
square root of a negative; huge operands overflowing to a non-finite result;
a negative base with a fractional exponent reaching the NaN branch; and the
float-precision cases documenting the rounding and percent-ordering decisions,
which use exact comparison rather than the shared epsilon tolerance.

**`internal/handlers`**, using `httptest.NewRecorder()` with no real server:
valid request returning 200 and the expected JSON body; malformed JSON, empty
body, trailing data, missing required field, unknown operation, non-numeric
operand and division by zero all returning 400 with a useful error code; wrong
method returning 405; unknown route returning 404; wrong `Content-Type`
returning 415; oversized body returning 413; and a direct test of the
error-mapping table for the 500 default path, which is unreachable over HTTP.

**`internal/httpx`**: preflight from the allowed origin returns 204 with the
CORS headers, a disallowed origin receives no `Access-Control-Allow-Origin`,
and a non-`OPTIONS` request reaches the next handler.

**`cmd/api`** is not tested. It wires dependencies and starts the server;
covering it would mean testing the standard library. The README says so.

**Frontend**, Vitest and React Testing Library, driving the real UI rather
than calling hooks directly, with `fetch` stubbed: a successful calculation
asserting both the request payload and the rendered result; an API error
rendering a message rather than crashing; keyboard entry updating the display;
validation blocking submission until an operation and both operands exist;
chaining computing when a second operator is pressed; operators, equals and
clear from the keyboard; and a network failure reported in readable terms.

Two cases considered and folded in: multiple decimal points in one operand
(`1.2.3`) cannot be typed, since the reducer ignores a second decimal point,
and non-numeric operands are rejected at the decoding layer and share the
`INVALID_BODY` code rather than having one of their own.

**Working method:** architecture, the error contract, the decisions in
section 4, scope, test case selection and every design sentence in the README
are mine — discuss them, do not generate them. Scaffolding, CORS middleware,
Dockerfiles and CSS can be generated for review. Handlers and calculation
logic only on request. Do not expand scope silently, do not add dependencies
without asking, prefer the standard library, and say plainly if something
here is wrong rather than working around it. Every prompt gets logged in
PROMPTS.md, including architecture discussion and rejected suggestions with
one line on why.

**Done when:** `gofmt -l .` silent and `go vet ./...` clean; `npm run lint`
and `npm run build` clean; tests passing with a committed coverage report;
`docker compose up` bringing both services up working; README covering setup,
how to run both layers, curl examples, design decisions with reasoning, and
what I would add with more time; PROMPTS.md complete; meaningful commit
history; `.gitignore` covering `node_modules`, build output, `.env` and
coverage files.

---

## 7. Decisions revised during implementation

Recorded rather than silently overwritten, since the reasoning matters more
than the first answer.

- **Pressing an operator with one pending** was going to replace it. It now
  computes when a second operand has already been typed, so `12 + 5 +` shows
  17, which is what a physical calculator does. Replacement still happens when
  no second operand has been entered.
- **Wrong `Content-Type`** was specified as 400. Corrected to 415, which RFC
  9110 defines for exactly this case. Oversized bodies likewise moved from 400
  to 413.
- **Rounding** was specified as a fixed number of decimal places. That
  overflows for large results and rounds small ones to zero, so it became
  significant digits.
- **A sign toggle** was not in the original scope and was added once it became
  clear the keypad had no way to enter a negative number.
