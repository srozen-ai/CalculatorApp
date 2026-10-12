import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import App from "./App";

/** mockFetch installs a fetch that resolves with the given status and body. */
function mockFetch(status: number, body: unknown) {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

/** display returns the result element, which is an <output> (role "status"). */
function display() {
  return screen.getByRole("status");
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("App", () => {
  it("sends the entered operation and renders the result", async () => {
    const fetchMock = mockFetch(200, { result: 17 });
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: "1" }));
    await user.click(screen.getByRole("button", { name: "2" }));
    await user.click(screen.getByRole("button", { name: "add" }));
    await user.click(screen.getByRole("button", { name: "5" }));
    await user.click(screen.getByRole("button", { name: "=" }));

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, init] = fetchMock.mock.calls[0];
    expect(JSON.parse(init.body as string)).toEqual({
      operation: "add",
      a: 12,
      b: 5,
    });
    expect(await screen.findByText("17")).toBeInTheDocument();
  });

  it("renders the API error message instead of crashing", async () => {
    mockFetch(400, {
      error: { code: "DIVISION_BY_ZERO", message: "division by zero" },
    });
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: "1" }));
    await user.click(screen.getByRole("button", { name: "divide" }));
    await user.click(screen.getByRole("button", { name: "0" }));
    await user.click(screen.getByRole("button", { name: "=" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "division by zero",
    );
  });

  it("updates the display from keyboard input", async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.keyboard("42");

    expect(screen.getByText("42")).toBeInTheDocument();
  });

  it("blocks submission until an operation and both operands exist", async () => {
    const fetchMock = mockFetch(200, { result: 0 });
    const user = userEvent.setup();
    render(<App />);

    const equals = screen.getByRole("button", { name: "=" });
    expect(equals).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "7" }));
    expect(equals).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "add" }));
    await user.click(screen.getByRole("button", { name: "3" }));
    expect(equals).toBeEnabled();

    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("computes the pending operation when a second operator is pressed", async () => {
    const fetchMock = mockFetch(200, { result: 17 });
    const user = userEvent.setup();
    render(<App />);

    await user.keyboard("12+5");
    await user.click(screen.getByRole("button", { name: "subtract" }));

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(await screen.findByText("17")).toBeInTheDocument();
  });

  it("supports operators, equals and clear from the keyboard", async () => {
    mockFetch(200, { result: 20 });
    const user = userEvent.setup();
    render(<App />);

    await user.keyboard("4*5{Enter}");
    expect(await screen.findByText("20")).toBeInTheDocument();

    await user.keyboard("{Escape}");
    expect(display()).toHaveTextContent("0");
  });

  it("reports a network failure in readable terms", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new TypeError("Failed to fetch")),
    );
    const user = userEvent.setup();
    render(<App />);

    await user.keyboard("9/3{Enter}");

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Could not reach the server",
    );
  });

  it("toggles the sign of the current operand", async () => {
    const fetchMock = mockFetch(200, { result: -2 });
    const user = userEvent.setup();
    render(<App />);

    await user.keyboard("5");
    await user.click(screen.getByRole("button", { name: "toggle sign" }));
    expect(display()).toHaveTextContent("-5");

    await user.keyboard("+3{Enter}");

    expect(JSON.parse(fetchMock.mock.calls[0][1].body as string)).toEqual({
      operation: "add",
      a: -5,
      b: 3,
    });
  });

  it("accepts one decimal point and deletes with backspace", async () => {
    const user = userEvent.setup();
    render(<App />);

    // The second decimal point is ignored, so "1.2.3" can never be formed.
    await user.keyboard("1.2.3");
    expect(display()).toHaveTextContent("1.23");

    await user.keyboard("{Backspace}");
    expect(display()).toHaveTextContent("1.2");
  });
});
