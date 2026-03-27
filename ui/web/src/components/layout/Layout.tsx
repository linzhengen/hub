import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Breadcrumb } from './Breadcrumb';
import { Button } from '@/components/ui/button';
import { LogOut, User, Menu, ChevronLeft, ChevronRight } from 'lucide-react';
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { useSidebar } from '@/hooks/use-sidebar';
import { useAuth } from '@/components/AuthProvider';

export function Layout() {
  const { logout } = useAuth();
  const { isCollapsed, isMobileDrawerOpen, toggleCollapsed, openMobileDrawer, closeMobileDrawer, setIsMobileDrawerOpen } = useSidebar();

  const handleLogout = () => {
    logout();
  };

  // Dynamic grid classes based on collapsed state
  const gridClass = isCollapsed
    ? 'md:grid-cols-[80px_1fr] lg:grid-cols-[80px_1fr]'
    : 'md:grid-cols-[220px_1fr] lg:grid-cols-[280px_1fr]';

  return (
    <Sheet open={isMobileDrawerOpen} onOpenChange={setIsMobileDrawerOpen}>
      <div className={`grid min-h-screen w-full ${gridClass}`}>
        {/* Desktop Sidebar */}
        <div className="hidden border-r bg-muted/40 md:block">
          <Sidebar collapsed={isCollapsed} />
        </div>

        <div className="flex flex-col">
          <header className="flex h-14 items-center gap-4 border-b bg-muted/40 px-4 lg:h-[60px] lg:px-6">
            {/* Mobile hamburger menu */}
            <SheetTrigger
              className="inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 hover:bg-muted hover:text-foreground aria-expanded:bg-muted aria-expanded:text-foreground dark:hover:bg-muted/50 size-8 flex md:hidden"
            >
              <Menu className="h-5 w-5" />
              <span className="sr-only">Toggle menu</span>
            </SheetTrigger>

            {/* Desktop sidebar toggle */}
            <Button
              variant="ghost"
              size="icon"
              className="hidden md:flex"
              onClick={toggleCollapsed}
            >
              {isCollapsed ? (
                <ChevronRight className="h-5 w-5" />
              ) : (
                <ChevronLeft className="h-5 w-5" />
              )}
              <span className="sr-only">Toggle sidebar</span>
            </Button>

            {/* Breadcrumb */}
            <div className="flex-1">
              <Breadcrumb />
            </div>

            {/* User dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger
                className="group/button inline-flex shrink-0 items-center justify-center rounded-lg border border-transparent bg-clip-padding text-sm font-medium whitespace-nowrap transition-all outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 active:not-aria-[haspopup]:translate-y-px disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 bg-secondary text-secondary-foreground hover:bg-secondary/80 aria-expanded:bg-secondary aria-expanded:text-secondary-foreground size-8 rounded-full"
              >
                <User className="h-5 w-5" />
                <span className="sr-only">Toggle user menu</span>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuGroup>
                  <DropdownMenuLabel>My Account</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={handleLogout}>
                    <LogOut className="mr-2 h-4 w-4" />
                    Logout
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </header>
          <main className="flex flex-1 flex-col gap-4 p-4 lg:gap-6 lg:p-6">
            <Outlet />
          </main>
        </div>
      </div>

      {/* Mobile Sheet Sidebar */}
      <SheetContent side="left" className="p-0 w-64">
        <Sidebar collapsed={false} />
      </SheetContent>
    </Sheet>
  );
}
