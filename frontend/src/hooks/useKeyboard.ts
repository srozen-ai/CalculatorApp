import { useEffect } from "react";

import type { BinaryOperation } from "../types/api";

const KEY_TO_OPERATION: Record<string, BinaryOperation> = {
  "+": "add",
  "-": "subtract",
  "*": "multiply",
  "/": "divide",
  "^": "exponent",
  "%": "percent",
};

type Handlers = {
  onDigit: (value: string) => void;
  onDecimal: () => void;
  onOperation: (operation: BinaryOperation) => void;
  onEquals: () => void;
  onClear: () => void;
  onBackspace: () => void;
  onToggleSign: () => void;
};

export function useKeyboard(handlers: Handlers) {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      // Leave browser and OS shortcuts alone.
      if (event.ctrlKey || event.metaKey || event.altKey) return;

      const target = event.target;
      const onButton =
        target instanceof HTMLElement && target.tagName === "BUTTON";

      if (event.key >= "0" && event.key <= "9") {
        handlers.onDigit(event.key);
        return;
      }
      if (event.key === "." || event.key === ",") {
        handlers.onDecimal();
        return;
      }
      if (event.key in KEY_TO_OPERATION) {
        // Firefox opens quick-find on "/".
        event.preventDefault();
        handlers.onOperation(KEY_TO_OPERATION[event.key]);
        return;
      }
      if (event.key === "Enter" || event.key === "=") {
        // With focus on a button, Enter already activates it; handling it here
        // too would fire two actions from one keypress.
        if (onButton) return;
        event.preventDefault();
        handlers.onEquals();
        return;
      }
      if (event.key === "Escape") {
        handlers.onClear();
        return;
      }
      if (event.key === "n" || event.key === "N") {
        handlers.onToggleSign();
        return;
      }
      if (event.key === "Backspace") {
        event.preventDefault();
        handlers.onBackspace();
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handlers]);
}
