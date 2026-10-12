export const BINARY_OPERATIONS = [
  "add",
  "subtract",
  "multiply",
  "divide",
  "exponent",
  "percent",
] as const;

export type BinaryOperation = (typeof BINARY_OPERATIONS)[number];
export type UnaryOperation = "sqrt";
export type Operation = BinaryOperation | UnaryOperation;

/**
 * A request is either a binary operation with both operands, or a unary one
 * with a single operand. The union makes an invalid combination — `add`
 * without `b`, or `sqrt` with one — a compile error rather than a 400.
 */
export type CalcRequest =
  | { operation: BinaryOperation; a: number; b: number }
  | { operation: UnaryOperation; a: number };

export type CalcSuccess = { result: number };

export type CalcError = {
  error: { code: string; message: string };
};
