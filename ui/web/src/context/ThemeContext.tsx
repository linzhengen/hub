import type React from "react";
import { useState, useEffect } from "react";
import { ThemeContext, type Theme } from "@/context/theme";

// 初期テーマは lazy initializer で読む。以前は effect 内で localStorage を読んで
// setState していたため、初回描画が必ず light になり保存済みテーマへ切り替わる
// までちらついていた（react-hooks/set-state-in-effect）。
const readStoredTheme = (): Theme => (localStorage.getItem("theme") as Theme | null) ?? "light";

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [theme, setTheme] = useState<Theme>(readStoredTheme);

  useEffect(() => {
    localStorage.setItem("theme", theme);
    document.documentElement.classList.toggle("dark", theme === "dark");
  }, [theme]);

  const toggleTheme = () => {
    setTheme((prevTheme) => (prevTheme === "light" ? "dark" : "light"));
  };

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
};
