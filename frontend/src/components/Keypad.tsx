import type { BinaryOperation } from "../types/api";
import { OPERATION_SYMBOLS } from "../operations";

type KeypadProps = {
  onDigit: (value: string) => void;
  onDecimal: () => void;
  onOperation: (operation: BinaryOperation) => void;
  onEquals: () => void;
  onSquareRoot: () => void;
  onClear: () => void;
  onBackspace: () => void;
  onToggleSign: () => void;
  canSubmit: boolean;
  loading: boolean;
};

/**
 * blurOnMouse releases focus after a mouse click so the button does not keep
 * swallowing later Enter presses. event.detail is 0 when the click came from
 * the keyboard, where focus must stay put.
 */
function blurOnMouse(event: React.MouseEvent<HTMLButtonElement>) {
  if (event.detail > 0) event.currentTarget.blur();
}

export function Keypad({
  onDigit,
  onDecimal,
  onOperation,
  onEquals,
  onSquareRoot,
  onClear,
  onBackspace,
  onToggleSign,
  canSubmit,
  loading,
}: KeypadProps) {
  const digit = (d: string) => (
    <button
      key={d}
      type="button"
      onClick={(e) => {
        blurOnMouse(e);
        onDigit(d);
      }}
    >
      {d}
    </button>
  );

  const operation = (op: BinaryOperation) => (
    <button
      key={op}
      type="button"
      className="key--operation"
      onClick={(e) => {
        blurOnMouse(e);
        onOperation(op);
      }}
      disabled={loading}
      aria-label={op}
    >
      {OPERATION_SYMBOLS[op]}
    </button>
  );

  return (
    <div className="keypad">
      <button
        type="button"
        className="key--util"
        onClick={(e) => {
          blurOnMouse(e);
          onClear();
        }}
      >
        C
      </button>
      <button
        type="button"
        className="key--util"
        onClick={(e) => {
          blurOnMouse(e);
          onBackspace();
        }}
        aria-label="backspace"
      >
        ⌫
      </button>
      <button
        type="button"
        className="key--util"
        onClick={(e) => {
          blurOnMouse(e);
          onSquareRoot();
        }}
        disabled={loading}
        aria-label="square root"
      >
        √
      </button>
      {operation("divide")}

      {["7", "8", "9"].map(digit)}
      {operation("multiply")}

      {["4", "5", "6"].map(digit)}
      {operation("subtract")}

      {["1", "2", "3"].map(digit)}
      {operation("add")}

      {operation("percent")}
      {digit("0")}
      <button
        type="button"
        onClick={(e) => {
          blurOnMouse(e);
          onDecimal();
        }}
      >
        .
      </button>
      {operation("exponent")}

      <button
        type="button"
        className="key--equals"
        onClick={(e) => {
          blurOnMouse(e);
          onEquals();
        }}
        disabled={!canSubmit}
      >
        =
      </button>
      <button
        type="button"
        className="key--util"
        onClick={(e) => {
          blurOnMouse(e);
          onToggleSign();
        }}
        aria-label="toggle sign"
      >
        ±
      </button>
    </div>
  );
}
