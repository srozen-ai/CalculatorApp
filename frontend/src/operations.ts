import type { BinaryOperation } from "./types/api";

export const OPERATION_SYMBOLS: Record<BinaryOperation, string> = {
  divide: "÷",
  multiply: "×",
  subtract: "−",
  add: "+",
  exponent: "^",
  percent: "%",
};
