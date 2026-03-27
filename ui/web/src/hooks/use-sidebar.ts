import { useState, useEffect } from 'react';

const SIDEBAR_COLLAPSED_KEY = 'sidebar-collapsed';

export function useSidebar() {
  const [isMobileDrawerOpen, setIsMobileDrawerOpen] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(() => {
    if (typeof window === 'undefined') return false;
    const saved = localStorage.getItem(SIDEBAR_COLLAPSED_KEY);
    return saved ? JSON.parse(saved) : false;
  });

  useEffect(() => {
    localStorage.setItem(SIDEBAR_COLLAPSED_KEY, JSON.stringify(isCollapsed));
  }, [isCollapsed]);

  const toggleCollapsed = () => setIsCollapsed(!isCollapsed);
  const openMobileDrawer = () => setIsMobileDrawerOpen(true);
  const closeMobileDrawer = () => setIsMobileDrawerOpen(false);

  return {
    isCollapsed,
    isMobileDrawerOpen,
    toggleCollapsed,
    openMobileDrawer,
    closeMobileDrawer,
    setIsCollapsed,
    setIsMobileDrawerOpen,
  };
}
