import { useMemo } from "react";

import { Display } from "./components/Display";
import { ErrorMessage } from "./components/ErrorMessage";
import { Keypad } from "./components/Keypad";
import { useCalculator } from "./hooks/useCalculator";
import { useKeyboard } from "./hooks/useKeyboard";
import "./App.css";

export default function App() {
  const calc = useCalculator();

  // Memoised so the keyboard effect does not re-subscribe on every render.
  const handlers = useMemo(
    () => ({
      onDigit: calc.inputDigit,
      onDecimal: calc.inputDecimal,
      onOperation: calc.chooseOperation,
      onEquals: calc.equals,
      onClear: calc.clear,
      onBackspace: calc.backspace,
      onToggleSign: calc.toggleSign,
    }),
    [
      calc.inputDigit,
      calc.inputDecimal,
      calc.chooseOperation,
      calc.equals,
      calc.clear,
      calc.backspace,
      calc.toggleSign,
    ],
  );

  useKeyboard(handlers);

  return (
    <main className="app">
      <h1 className="app__title">Calculator</h1>

      <section className="calculator" aria-label="calculator">
        <Display
          value={calc.display}
          pendingOperand={calc.pendingOperand}
          pendingOperation={calc.pendingOperation}
          awaitingOperand={calc.awaitingOperand}
          loading={calc.loading}
        />
        <ErrorMessage message={calc.error} />
        <Keypad
          onDigit={calc.inputDigit}
          onDecimal={calc.inputDecimal}
          onOperation={calc.chooseOperation}
          onEquals={calc.equals}
          onSquareRoot={calc.squareRoot}
          onClear={calc.clear}
          onBackspace={calc.backspace}
          onToggleSign={calc.toggleSign}
          canSubmit={calc.canSubmit}
          loading={calc.loading}
        />
      </section>
    </main>
  );
}
