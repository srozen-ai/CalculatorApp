import { useCallback, useReducer } from "react";

import { calculate } from "../api/calculate";
import { ApiError } from "../api/errors";
import type { BinaryOperation } from "../types/api";

type State = {
  /** The operand being typed, kept as a string so "1." and "0.0" are representable. */
  display: string;
  pendingOperand: number | null;
  pendingOperation: BinaryOperation | null;
  /** When true, the next digit replaces the display instead of appending. */
  overwrite: boolean;
  error: string | null;
  loading: boolean;
};

type Action =
  | { type: "digit"; value: string }
  | { type: "decimal" }
  | { type: "backspace" }
  | { type: "clear" }
  | { type: "operation"; operation: BinaryOperation }
  | { type: "submitStart" }
  | { type: "submitSuccess"; result: number }
  | { type: "toggleSign" }
  | { type: "submitFailure"; message: string };

const initialState: State = {
  display: "0",
  pendingOperand: null,
  pendingOperation: null,
  overwrite: false,
  error: null,
  loading: false,
};

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case "digit": {
      const display =
        state.overwrite || state.display === "0"
          ? action.value
          : state.display === "-0"
            ? `-${action.value}`
            : state.display + action.value;
      return { ...state, display, overwrite: false, error: null };
    }

    case "decimal": {
      if (state.overwrite) {
        return { ...state, display: "0.", overwrite: false, error: null };
      }
      // A second decimal point is ignored, so "1.2.3" can never be typed.
      if (state.display.includes(".")) return state;
      return { ...state, display: state.display + ".", error: null };
    }

    case "backspace": {
      if (state.overwrite) return { ...state, display: "0", overwrite: false };
      const display = state.display.slice(0, -1);
      return { ...state, display: display === "" ? "0" : display };
    }

    case "clear":
      return initialState;

    case "operation": {
      const operand = Number(state.display);
      if (!Number.isFinite(operand)) {
        return { ...state, error: "Enter a number first" };
      }
      // Stores the current display as the first operand. Chaining — computing
      // when a second operand is already present — is handled in the hook,
      // which needs to call the API; this transition stays pure.
      return {
        ...state,
        pendingOperand: operand,
        pendingOperation: action.operation,
        overwrite: true,
        error: null,
      };
    }

    case "submitStart":
      return { ...state, loading: true, error: null };

    case "submitSuccess":
      return {
        ...state,
        display: String(action.result),
        pendingOperand: null,
        pendingOperation: null,
        overwrite: true,
        loading: false,
        error: null,
      };

    case "submitFailure":
      return { ...state, loading: false, error: action.message };

    case "toggleSign": {
      // Pressing ± after choosing an operator starts the second operand as a
      // negative number rather than negating the one already shown.
      if (state.overwrite) {
        return { ...state, display: "-0", overwrite: false, error: null };
      }
      const display = state.display.startsWith("-")
        ? state.display.slice(1)
        : `-${state.display}`;
      return { ...state, display, error: null };
    }
  }
}

export function useCalculator() {
  const [state, dispatch] = useReducer(reducer, initialState);

  const run = useCallback(async (request: Parameters<typeof calculate>[0]) => {
    dispatch({ type: "submitStart" });
    try {
      dispatch({ type: "submitSuccess", result: await calculate(request) });
    } catch (err) {
      dispatch({
        type: "submitFailure",
        message: err instanceof ApiError ? err.message : "Something went wrong",
      });
    }
  }, []);

  const canSubmit =
    !state.loading &&
    state.pendingOperation !== null &&
    state.pendingOperand !== null &&
    Number.isFinite(Number(state.display));

  const equals = useCallback(() => {
    if (!canSubmit) return;
    void run({
      operation: state.pendingOperation as BinaryOperation,
      a: state.pendingOperand as number,
      b: Number(state.display),
    });
  }, [
    canSubmit,
    run,
    state.pendingOperand,
    state.pendingOperation,
    state.display,
  ]);

  const squareRoot = useCallback(() => {
    const a = Number(state.display);
    if (state.loading || !Number.isFinite(a)) return;
    void run({ operation: "sqrt", a });
  }, [run, state.display, state.loading]);

  // Wrapped in useCallback so their identity is stable across renders, which
  // keeps the keyboard listener in App from re-subscribing on every keystroke.
  const inputDigit = useCallback(
    (value: string) => dispatch({ type: "digit", value }),
    [],
  );
  const inputDecimal = useCallback(() => dispatch({ type: "decimal" }), []);
  const backspace = useCallback(() => dispatch({ type: "backspace" }), []);
  const clear = useCallback(() => dispatch({ type: "clear" }), []);
  const toggleSign = useCallback(() => dispatch({ type: "toggleSign" }), []);

  /**
   * chooseOperation computes the pending operation first when a second operand
   * has already been typed, so "12 + 5 +" shows 17 and chains from there.
   * Pressing an operator before typing an operand replaces the pending one.
   */
  const chooseOperation = useCallback(
    (operation: BinaryOperation) => {
      const hasSecondOperand =
        state.pendingOperation !== null &&
        state.pendingOperand !== null &&
        !state.overwrite;

      if (state.loading || !hasSecondOperand) {
        dispatch({ type: "operation", operation });
        return;
      }

      void (async () => {
        dispatch({ type: "submitStart" });
        try {
          const result = await calculate({
            operation: state.pendingOperation as BinaryOperation,
            a: state.pendingOperand as number,
            b: Number(state.display),
          });
          // The reducer applies these in order, so "operation" sees the result
          // as the new display and stores it as the next first operand.
          dispatch({ type: "submitSuccess", result });
          dispatch({ type: "operation", operation });
        } catch (err) {
          dispatch({
            type: "submitFailure",
            message:
              err instanceof ApiError ? err.message : "Something went wrong",
          });
        }
      })();
    },
    [
      state.loading,
      state.overwrite,
      state.pendingOperand,
      state.pendingOperation,
      state.display,
    ],
  );

  return {
    display: state.display,
    pendingOperand: state.pendingOperand,
    pendingOperation: state.pendingOperation,
    /** True while waiting for the second operand to be typed. */
    awaitingOperand: state.overwrite,
    error: state.error,
    loading: state.loading,
    canSubmit,
    inputDigit,
    inputDecimal,
    backspace,
    clear,
    toggleSign,
    chooseOperation,
    equals,
    squareRoot,
  };
}
