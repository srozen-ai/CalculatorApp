import type { BinaryOperation } from "../types/api";
import { OPERATION_SYMBOLS } from "../operations";

type DisplayProps = {
  value: string;
  pendingOperand: number | null;
  pendingOperation: BinaryOperation | null;
  awaitingOperand: boolean;
  loading: boolean;
};

/**
 * scaleFont shrinks the result as it grows so long numbers stay on one line
 * instead of wrapping or forcing a scrollbar.
 */
function scaleFont(length: number): string {
  if (length <= 10) return "2.5rem";
  const size = 2.5 - (length - 10) * 0.13;
  return `${Math.max(1, size)}rem`;
}

export function Display({
  value,
  pendingOperand,
  pendingOperation,
  awaitingOperand,
  loading,
}: DisplayProps) {
  // While awaiting the second operand the display still holds the first one,
  // so showing it would render "12 × 12".
  const expression =
    pendingOperand === null || pendingOperation === null
      ? ""
      : awaitingOperand
        ? `${pendingOperand} ${OPERATION_SYMBOLS[pendingOperation]}`
        : `${pendingOperand} ${OPERATION_SYMBOLS[pendingOperation]} ${value}`;

  return (
    <div className="display">
      <div className="display__expression" aria-hidden="true">
        {expression}
      </div>
      <output
        className="display__value"
        style={{ fontSize: scaleFont(value.length) }}
        aria-live="polite"
        aria-busy={loading}
      >
        {value}
      </output>
    </div>
  );
}
