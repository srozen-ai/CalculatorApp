# Prompts

How this project was built with AI assistance, as the assignment requested.

**On accuracy:** these prompts are reconstructed from the conversations after
the fact rather than captured verbatim. The substance and sequence are accurate;
the wording of my messages has been cleaned up.

**Tooling:** Claude, in two parallel conversations. One was the main working
session — architecture discussion, code generation and debugging. The second was
used only as a reviewer: I pasted finished files and asked for a critique against
the project's own conventions, without letting it write replacements. Several of
the sharpest findings came from that second conversation, including one that
corrected an answer the first had given me.

**What I did not delegate:** project structure and package boundaries, the error
contract and status code mapping, the floating-point and percent decisions,
scope, test case selection, and the design reasoning in the README. Those were
decided first and written into CONTEXT.md, which was pasted at the start of the
working session so generated code had to conform to them.

---

## 1. Setting the brief

Before writing code I wrote CONTEXT.md — the assignment, scope, architecture,
the API contract, style authorities, and the test plan — and used it as the
standing context for every later prompt.

> Here is the assignment email. Help me write a context file for AI assistance
> on this project: what is being built, what is out of scope, the architecture,
> the decisions that need justifying, and how I want to work. Keep it to about
> five sections, non-verbose, and add an instruction that every answer briefly
> explains the output it gave.

The first draft came back long and editorialising, with judgments in it that do
not belong in a repository file. I had it rewritten in prose at roughly half the
length, with exponentiation moved into scope and the one-endpoint rationale
restructured to name the alternative first, then what it gains and what it gives
up.

---

## 2. Backend — the arithmetic layer

I wrote `internal/calc` by hand and used AI for review rather than generation.

> Here is CONTEXT.md. I've written `internal/calc` myself — the pure arithmetic
> layer with its sentinel errors. Review it against the Go conventions in
> section 5 and list what it gets wrong, ordered by impact. Don't rewrite it.
> Also tell me whether my error cases cover the branches in section 6, and name
> any branch I've missed.

This caught that my four sentinel errors were declared and never used — `Divide`
returned `+Inf` rather than `ErrDivisionByZero`, and `Sqrt(-4)` returned `NaN`.

> What happens if b isn't passed? Does that surface as ErrNotFinite, or does it
> need its own error?

`calc` never sees a missing operand — that distinction belongs to the request
type in the handler layer, which is what led to `a` and `b` being pointers.

> Help me write test cases for arithmetic.go. One is enough and I'll write the
> rest.

I wrote the rest from the pattern given. Later, on reviewing the file:

> Is there a way to reduce the file size for the testing file?

The six binary operations shared an identical 14-line loop, extracted into a
`runBinary` helper. The same was proposed for `TestSqrt` via a second helper;
**I rejected that one** — a helper with a single caller would have cost about as
many lines as it saved.

---

## 3. Backend — the HTTP layer

> Help me fix the errors here: [pasted a draft that mixed HTTP plumbing,
>
> > operation dispatch and status mapping in one function]

The correction was to split those three jobs into separate files, which is the
structure the project kept.

> Where should the status mapping go — its own file inside handlers? And where
> do the handler tests go, structured how?

> What is missing here? [pasted calculate_http.go]

Then a round of manual verification with curl against every error branch, which
is what surfaced the next item.

---

## 4. The review conversation

At several points I pasted finished files into the second conversation:

> Review this against the conventions in CONTEXT.md and list what an
> experienced reviewer would flag, ordered by impact. Don't rewrite it.

Findings worth recording, because they changed the code:

- **The float-precision test could not fail.** With an epsilon tolerance of
  `1e-9`, asserting `0.30000000000000004` would also pass for `0.3`, defeating
  the purpose. That case now uses exact comparison, and the epsilon comment
  states the exception.
- **The `0^0` justification cited the wrong authority.** The contract actually
  relied on is `math.Pow`'s documented guarantee that `Pow(x, ±0) = 1`; IEEE
  754's `powr(0, 0)` is NaN. The README cites `math.Pow`.
- **`statusFor` and `codeFor` were two switches that had to stay in sync.**
  Replaced by a single table walked with `errors.Is`.
- **`writeJSON` wrote the status before encoding**, so a marshaling failure
  could follow a 200. Now it marshals first.
- **Wrong `Content-Type` belongs at 415, not 400**, and an oversized body at 413. CONTEXT.md was corrected to match.
- **A missing `a` was silently treated as zero**, because only `b` was a
  pointer.
- **Rounding to fixed decimal places failed at both ends** — `1 ÷ 1e12` rounded
  to `0`, and large results overflowed. Changed to significant digits.

Suggestions I **did not** take: a multi-origin CORS allowlist, and rejecting
unknown JSON fields. Both are defensible either way; the reasoning for declining
is in the README.

---

## 5. A bug found by prompting for one

> Is there anything wrong with this handler? [pasted calculate_http.go]

The answer predicted that `resultPrecision = 10` would overflow for large
results, and gave a curl to prove it:

```bash
curl -s -X POST localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"multiply","a":1e300,"b":10}'
```

It returned an empty body with `json: unsupported value: +Inf` in the log, after
a 200 had already been written. This is the single most useful thing AI did on
this project, and it is why the rounding is to significant digits and why
`writeJSON` marshals before setting the status.

---

## 6. Frontend

> I'm about to build the frontend. Two questions. Where should input validation
> live — in the hook, or in the components? And since the API takes discrete
> operands rather than an expression, a calculator UI where you type `12 + 5 =`
> has to hold "pending operand plus pending operation" state locally. Is there a
> cleaner way to model that than a reducer, given the scope?

The answer reframed the second question: the reducer only matters if the UI is a
keypad rather than a two-field form, and at four coupled fields a reducer is the
right tool rather than over-engineering. I chose the keypad.

> Help me with the hook.

Two UI decisions came out of the next round. Pressing the _next_ operator now
computes, which is what physical calculators do and which reversed my earlier
decision to replace the pending operator — a calculator cannot know when typing
has finished, so it is the only point at which the intent is unambiguous. And
`percent` stayed binary: physical calculators treat `%` as a contextual modifier
reading the preceding operand, which this API has no expression context to
support, so the difference is documented rather than imitated.

> Sometimes Enter and Backspace break. Why, and how would I fix it?

The cause was a guard I had added to prevent Enter firing twice: after clicking
a button with the mouse, that button keeps focus, so every later keypress was
being discarded. Fixed by releasing focus on mouse-initiated clicks only
(`event.detail > 0`) and dropping the guard for Backspace, which cannot activate
a button.

> There's no button for negative values — the minus only subtracts.

> Set up Vitest and React Testing Library, and write the first test: a
> successful calculation with fetch stubbed, asserting the request payload as
> well as the rendered result.

> [pasted a failing test] Found multiple elements with the text: 0

The display and the `0` key both rendered "0". Fixed by querying the `<output>`
by its implicit `status` role rather than by text.

---

## 7. Docker and documentation

> The backend and frontend both work. Help me write the Dockerfiles and
> docker-compose.

The notes that mattered here were that `VITE_API_URL` is baked in at build time
so it has to be a build arg, that it must be an address the _browser_ can
resolve rather than the Compose service name, and that `CGO_ENABLED=0` is what
makes the distroless runtime image possible.

> Let's write the README with what the email asked for, the correct formatting,
> and everything that might be important given the context and the current
> repository.

A second pass on the README added the `internal/httpx` tests, two more frontend
tests, curl examples for non-numeric operands and a missing `Content-Type`, the
project structure as an annotated tree, and the explicit note that per-operation
routes were considered and rejected rather than overlooked.
