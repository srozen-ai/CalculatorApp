# Calculator

A full-stack calculator: a React + TypeScript frontend consuming a Go REST API
that performs the arithmetic. The backend is a separate, independently
deployable service; the frontend holds only input state and talks to it over
HTTP/JSON.

---

## Contents

- [Quick start](#quick-start)
- [Running without Docker](#running-without-docker)
- [Configuration](#configuration)
- [API](#api)
- [Tests and coverage](#tests-and-coverage)
- [Project structure](#project-structure)
- [Design decisions](#design-decisions)
- [Known limitations](#known-limitations)

---

## Quick start

Requires Docker with Compose.

```bash
docker compose up --build
```

The app is at <http://localhost:5173>, the API at <http://localhost:8080>.

To reach it from another device on the same network, pass your machine's LAN
address so the bundle is built against a URL the browser can resolve:

```bash
IP=$(ip route get 1.1.1.1 | awk '{print $7; exit}')
API_URL=http://$IP:8080 ALLOWED_ORIGIN=http://$IP:5173 docker compose up --build
```

---

## Running without Docker

Requires Go 1.26.8+ and Node 22+.

**Backend:**

```bash
cd backend
go run ./cmd/api
```

Listens on `:8080`.

**Frontend**, in a second terminal:

```bash
cd frontend
npm install
npm run dev
```

Serves on <http://localhost:5173>, which is the origin the backend allows by
default.

---

## Configuration

| Variable         | Service  | Default                 | Notes                                                                                          |
| ---------------- | -------- | ----------------------- | ---------------------------------------------------------------------------------------------- |
| `ADDR`           | backend  | `:8080`                 | Listen address. Binds all interfaces so containers and LAN devices can reach it.               |
| `ALLOWED_ORIGIN` | backend  | `http://localhost:5173` | Origin permitted by CORS. Must match where the browser loads the frontend from.                |
| `VITE_API_URL`   | frontend | `http://localhost:8080` | API base URL. Baked in at **build** time, so it is a Docker build arg, not a runtime variable. |

`VITE_API_URL` must be a URL the _browser_ can reach. Inside Compose the
backend is also reachable as `http://backend:8080`, but the request is made by
the browser, which is outside the Compose network, so that address would fail.

---

## API

One endpoint handles every operation.

### `POST /api/calculate`

Requires `Content-Type: application/json`.

```jsonc
{
  "operation": "add", // add | subtract | multiply | divide | exponent | percent | sqrt
  "a": 12.5,
  "b": 3, // omitted for sqrt
}
```

Success — `200`:

```json
{ "result": 15.5 }
```

Failure — one envelope for every error:

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "division by zero" } }
```

### Status codes

| Status | When                                                                                                                                                      |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `200`  | Success                                                                                                                                                   |
| `400`  | Malformed or empty JSON, trailing data, unknown operation, missing or non-numeric operand, division by zero, square root of a negative, non-finite result |
| `404`  | Unknown path                                                                                                                                              |
| `405`  | Wrong method on a known path                                                                                                                              |
| `413`  | Request body over 1 MiB                                                                                                                                   |
| `415`  | `Content-Type` is missing or not `application/json`                                                                                                       |
| `500`  | Unexpected server fault                                                                                                                                   |

### Error codes

`UNKNOWN_OPERATION` · `MISSING_OPERAND` · `DIVISION_BY_ZERO` ·
`NEGATIVE_ROOT` · `ZERO_NEGATIVE_POWER` · `NOT_FINITE` · `INVALID_BODY` ·
`BODY_TOO_LARGE` · `UNSUPPORTED_MEDIA_TYPE` · `INTERNAL_ERROR`

Codes are stable identifiers a client can branch on; messages are for humans
and may change.

### Examples

```bash
# Addition — note the rounding at the response boundary
curl -s -X POST localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":0.1,"b":0.2}'
# {"result":0.3}

# Square root takes a single operand
curl -s -X POST localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","a":9}'
# {"result":3}

# Percent is a% of b
curl -s -X POST localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"percent","a":50,"b":200}'
# {"result":100}

# Division by zero is a client error, not a server fault
curl -s -X POST localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"divide","a":1,"b":0}'
# {"error":{"code":"DIVISION_BY_ZERO","message":"division by zero"}}

# A missing operand is rejected; an operand of zero is not
curl -s -X POST localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":1}'
# {"error":{"code":"MISSING_OPERAND","message":"a required operand is missing"}}

# A non-numeric operand fails to decode and is rejected before any calculation
curl -s -X POST localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":"abc","b":2}'
# {"error":{"code":"INVALID_BODY","message":"request body must be valid JSON"}}

# A request without the JSON content type is rejected before the body is read
curl -s -X POST localhost:8080/api/calculate \
  -d '{"operation":"add","a":1,"b":2}'
# {"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json"}}
```

Operands that are not numbers — `"abc"`, or a string containing more than one
decimal point — are rejected by the JSON decoder, so they share the
`INVALID_BODY` code with a malformed body rather than having one of their own.
That is deliberate: the decoder cannot distinguish "this field was the wrong
type" from "this document was not valid JSON" without parsing twice.

---

## Tests and coverage

**Backend:**

```bash
cd backend
go test ./... -cover
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

| Package             | Coverage |
| ------------------- | -------- |
| `internal/calc`     | 100.0%   |
| `internal/handlers` | 84.1%    |
| `internal/httpx`    | 90.9%    |
| `cmd/api`           | 0.0%     |

**Frontend:**

```bash
cd frontend
npm run test
npm run test:coverage
```

| Metric     | Coverage |
| ---------- | -------- |
| Statements | 82.27%   |
| Branches   | 79.09%   |
| Functions  | 86.48%   |
| Lines      | 84.02%   |

9 tests, all driving the rendered UI rather than calling hooks directly.

### What is tested, and what is not

`internal/calc` is pure arithmetic with no HTTP, so every operation and every
error branch is covered, including overflow, NaN, and the float-precision cases
that document the rounding and percent-ordering decisions.

`internal/handlers` is tested through the router with `httptest`, so the 404 and
405 that the standard library's mux produces are exercised alongside the
handler's own branches. The error-to-status table is also tested directly, to
reach the 500 default path that is unreachable over HTTP. The uncovered
remainder is defensive: encoder failures and error-writing fallbacks.

`internal/httpx` covers the preflight response and the pass-through to the next
handler.

`cmd/api` has no tests. It wires dependencies and starts the server; covering it
would mean testing the standard library.

On the frontend, `src/types/api.ts` reports 0% because it is almost entirely
type declarations, which TypeScript erases at compile time — there is no runtime
code to execute. The uncovered remainder in `Keypad.tsx` is the JSX for keys no
test presses, and in `useCalculator.ts` the error branch of the chaining path.

Responsive behaviour was verified **manually** at 375px in browser devtools and
on a physical phone over the local network. It is not covered by automated tests
— jsdom does not lay out a viewport.

---

## Project structure

```
CalculatorApp/
├── backend/
│   ├── cmd/api/
│   │   └── main.go                  wiring, timeouts, startup
│   ├── internal/
│   │   ├── calc/                    pure arithmetic — no HTTP
│   │   │   ├── arithmetic.go        the seven operations + sentinel errors
│   │   │   └── arithmetic_test.go
│   │   ├── handlers/                HTTP layer
│   │   │   ├── calculate.go         request type, operand validation, dispatch
│   │   │   ├── calculate_http.go    decoding, rounding, the handler itself
│   │   │   ├── response.go          JSON and error writers
│   │   │   ├── router.go            route registration
│   │   │   ├── status.go            error → status + code table
│   │   │   └── calculate_http_test.go
│   │   └── httpx/
│   │       ├── cors.go              CORS middleware
│   │       └── cors_test.go
│   ├── Dockerfile                   multi-stage, distroless runtime
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   │   ├── calculate.ts         fetch layer, response validation
│   │   │   └── errors.ts            ApiError
│   │   ├── components/              Display, Keypad, ErrorMessage
│   │   ├── hooks/
│   │   │   ├── useCalculator.ts     reducer + async submission
│   │   │   └── useKeyboard.ts       global key handling
│   │   ├── types/api.ts             the request/response contract
│   │   ├── operations.ts            operation → symbol
│   │   └── App.test.tsx
│   ├── nginx.conf
│   ├── Dockerfile                   build with Vite, serve with nginx
│   └── package.json
├── docker-compose.yml
├── CONTEXT.md                       the brief this was built against
├── PROMPTS.md
└── README.md
```

`internal/` is enforced by the Go compiler against imports from outside the
module, and `cmd/` separates the binary entrypoint from library code. These two
directories come from common Go layout conventions;
[golang-standards/project-layout](https://github.com/golang-standards/project-layout)
is not an official standard and has been publicly criticised by Go maintainers,
so only the parts that apply at this scale were adopted.

---

## Design decisions

### One endpoint rather than one route per operation

The alternative was `POST /api/add`, `POST /api/divide`, and so on. A single
endpoint means request decoding, validation and error mapping are written once
instead of repeated in every handler, and adding an operation touches one switch
and one test table. The cost is that the operation travels in the body rather
than the URL path, which is less conventionally RESTful. Both designs are
defensible; this one was chosen for the reduction in duplicated validation.
Per-operation routes were considered and rejected for that reason, not
overlooked.

### `calc` knows nothing about HTTP

Arithmetic is pure functions returning `(float64, error)`. This is the
load-bearing boundary in the codebase: it makes the core testable without a
server, and it means a change to the transport cannot break the arithmetic.

### Floating point: compute in `float64`, round at the boundary

`0.1 + 0.2` is `0.30000000000000004` in IEEE-754. `calc` returns that value
unrounded, and rounding happens once, in the handler, immediately before
encoding. Rounding is to **15 significant digits** rather than a fixed number of
decimal places: fixed decimals destroy small results (`1 ÷ 1e12` would round to
`0`) and overflow on large ones (multiplying `1e301` by `1e10` yields `+Inf`,
which `json.Marshal` rejects). Fifteen digits is what `float64` represents
reliably.

Decimal or rational arithmetic was considered and rejected as overkill here —
though it would be mandatory if these were money values.

### Non-finite results are rejected, not encoded

`json.Marshal` fails on `Inf` and `NaN`. Without a guard, an overflow would send
a 200 status followed by an encoding failure and an empty body. `calc` routes
every result through a finiteness check, and `writeJSON` marshals before writing
the status so an encoding failure can never follow a success code.

### Division by zero is a 400

The input is what is wrong, not the server. The same applies to square roots of
negatives and to results that overflow: all are predictable, client-caused
conditions.

### Absent operands are distinguished from zero

`a` and `b` are `*float64` in the request struct. With plain `float64`, an
omitted operand would decode to `0` and `{"operation":"add","a":1}` would
silently return `1`. The pointer makes the missing case detectable, and it is
rejected with `MISSING_OPERAND`.

### `percent` is binary, not the physical-calculator modifier

A physical calculator computes `200 + 10%` as 220, because `%` reads the
preceding operand. That is a contextual modifier, not an operator. This API
takes discrete operands and has no expression context, so reproducing it would
require the client to chain two requests per keypress. `percent` is modeled as
`a% of b` instead, keeping every operation uniform.

`(a / 100) * b` is used rather than `a * b / 100` so that a large `b` cannot
overflow in an intermediate step. The cost is sub-epsilon error on common inputs
— 7% of 100 is `7.000000000000001` — which the response rounding absorbs. A test
pins this exact value so the behaviour is deliberate rather than accidental.

### `Content-Type` is enforced

A request without `application/json` is rejected with 415 before the body is
read. Beyond correctness, this makes every browser request non-simple under
CORS, so a cross-site `text/plain` POST cannot reach the handler without passing
the preflight.

### Validation lives on both layers

The frontend disables submission until an operation and both operands exist,
which is a usability concern. The backend validates independently and is
authoritative; it never trusts the client.

### Frontend state is a reducer

The keypad holds four fields that must change together — the display string, the
pending operand, the pending operation, and whether the next digit replaces or
appends. Separate `useState` calls allow intermediate states that should not
exist, which is where the classic calculator bugs come from. The reducer is
pure; all asynchrony lives in the hook, so every transition is testable without
rendering.

The display is kept as a **string**, not a number, because `"1."` and `"0.0"` are
valid mid-typing states a `number` cannot represent. Ignoring a second decimal
point at input time also means `1.2.3` can never be formed.

### Pressing an operator with one pending computes

`12 + 5 +` shows 17 and chains from there, which is what a physical calculator
does. Pressing an operator _before_ typing a second operand replaces the pending
one instead. The signal distinguishing the two is whether a digit has been
entered since the operator was chosen.

### No third-party backend dependencies

Routing uses `net/http` with Go 1.22+ method patterns
(`mux.HandleFunc("POST /api/calculate", …)`), which yields 405 and 404 without
extra code. Nothing to keep updated, and a smaller surface to audit.

### Other defensive choices

`http.MaxBytesReader` caps the request body at 1 MiB. The server sets read, write
and header timeouts rather than using `http.ListenAndServe` bare, which would
leave it open to slow-request attacks. The 500 path logs the real error and
returns a generic message, so internal detail is debuggable without being
disclosed.

---

## Known limitations

These were deliberate scope decisions, not oversights.

- **404 and 405 return plain text**, not the JSON error envelope. They are
  produced by the standard library's router before any handler runs; wrapping
  them would require intercepting the mux's own response writer.
- **Unknown JSON fields are accepted.** Rejecting them makes the API brittle to
  clients that add metadata; trailing data after the object _is_ rejected.
- **No calculation history, memory keys, or parenthesised expressions.** The API
  takes discrete operands by design, so an expression parser was out of scope.
- **`b` is ignored for `sqrt`** rather than rejected, since the request type on
  the client makes the invalid shape unrepresentable at compile time.
- **Results are rounded to 15 significant digits**, so a calculation needing more
  precision than `float64` provides is not supported.
- **CORS allows a single origin**, read from the environment at startup. A
  multi-origin allowlist would need the middleware to compare against the
  request's `Origin` header.

### With more time

Structured logging with request IDs; a `/healthz` endpoint for container health
checks; graceful shutdown on SIGTERM; a rate limiter; and property-based tests
for the arithmetic layer.
