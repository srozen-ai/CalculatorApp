import type { CalcRequest, CalcError, CalcSuccess } from "../types/api";
import { ApiError } from "./errors";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

function isCalcSuccess(value: unknown): value is CalcSuccess {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as CalcSuccess).result === "number"
  );
}

function isCalcError(value: unknown): value is CalcError {
  const body = value as CalcError;
  return (
    typeof body === "object" &&
    body !== null &&
    typeof body.error === "object" &&
    body.error !== null &&
    typeof body.error.message === "string" &&
    typeof body.error.code === "string"
  );
}

/**
 * calculate sends one operation to the API and returns its result.
 * Throws ApiError for anything the user should see.
 */
export async function calculate(req: CalcRequest): Promise<number> {
  let res: Response;
  try {
    res = await fetch(`${BASE_URL}/api/calculate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });
  } catch {
    // fetch rejects with TypeError for network failures and CORS rejections,
    // whose messages are not meaningful to a user.
    throw new ApiError("Could not reach the server", "NETWORK_ERROR", 0);
  }

  let body: unknown;
  try {
    body = await res.json();
  } catch {
    // Proxies and the router's own 404/405 return non-JSON bodies.
    throw new ApiError(
      `Unexpected response from the server (${res.status})`,
      "INVALID_RESPONSE",
      res.status,
    );
  }

  if (!res.ok) {
    if (isCalcError(body)) {
      throw new ApiError(body.error.message, body.error.code, res.status);
    }
    throw new ApiError(
      `Request failed (${res.status})`,
      "INVALID_RESPONSE",
      res.status,
    );
  }

  if (!isCalcSuccess(body)) {
    throw new ApiError(
      "Unexpected response from the server",
      "INVALID_RESPONSE",
      res.status,
    );
  }

  return body.result;
}
