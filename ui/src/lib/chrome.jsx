import { createContext, useContext } from "react";

const ChromeContext = createContext(null);

export function ChromeProvider({ value, children }) {
  return <ChromeContext.Provider value={value}>{children}</ChromeContext.Provider>;
}

export function useChrome() {
  const value = useContext(ChromeContext);
  if (!value) {
    throw new Error("useChrome must be used inside ChromeProvider");
  }

  return value;
}
