import { createContext, useContext } from "react";

export type SidebarContextType = {
  isExpanded: boolean;
  isMobileOpen: boolean;
  isHovered: boolean;
  activeItem: string | null;
  openSubmenu: string | null;
  toggleSidebar: () => void;
  toggleMobileSidebar: () => void;
  setIsHovered: (isHovered: boolean) => void;
  setActiveItem: (item: string | null) => void;
  toggleSubmenu: (item: string) => void;
};

export const SidebarContext = createContext<SidebarContextType | undefined>(undefined);

// Provider と別ファイルにしているのは、コンポーネント以外を同居させると
// Fast Refresh が効かなくなるため（react-refresh/only-export-components）。
export const useSidebar = () => {
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebar must be used within a SidebarProvider");
  }
  return context;
};
