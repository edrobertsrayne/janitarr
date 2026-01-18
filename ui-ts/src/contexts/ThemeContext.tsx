/**
 * Theme context for managing light/dark mode
 */

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useMemo,
} from "react";
import { ThemeProvider as MuiThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import { lightTheme, darkTheme } from "../theme";
import type { ThemeMode } from "../types";

interface ThemeContextValue {
  mode: ThemeMode;
  setMode: (mode: ThemeMode) => void;
  effectiveMode: "light" | "dark";
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

interface ThemeProviderProps {
  children: React.ReactNode;
}

/**
 * Provides theme management functionality to the app
 */
export function ThemeProvider({ children }: ThemeProviderProps) {
  // Load theme preference from localStorage, default to 'system'
  const [mode, setModeState] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem("janitarr-theme-mode");
    return (stored as ThemeMode) || "system";
  });

  // Determine if system prefers dark mode
  const [systemPrefersDark, setSystemPrefersDark] = useState(
    () => window.matchMedia("(prefers-color-scheme: dark)").matches,
  );

  // Listen for system theme changes
  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = (e: MediaQueryListEvent) => {
      setSystemPrefersDark(e.matches);
    };

    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, []);

  // Calculate effective theme mode
  const effectiveMode: "light" | "dark" = useMemo(() => {
    if (mode === "system") {
      return systemPrefersDark ? "dark" : "light";
    }
    return mode;
  }, [mode, systemPrefersDark]);

  // Update localStorage when mode changes
  const setMode = (newMode: ThemeMode) => {
    setModeState(newMode);
    localStorage.setItem("janitarr-theme-mode", newMode);
  };

  const theme = effectiveMode === "dark" ? darkTheme : lightTheme;

  const contextValue: ThemeContextValue = {
    mode,
    setMode,
    effectiveMode,
  };

  return (
    <ThemeContext.Provider value={contextValue}>
      <MuiThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </MuiThemeProvider>
    </ThemeContext.Provider>
  );
}

/**
 * Hook to access theme context
 */
export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within ThemeProvider");
  }
  return context;
}
