import { Link, useLocation } from 'react-router-dom';
import { Users, Shield, Key, FileText, LayoutDashboard, FolderKanban } from 'lucide-react';
import { cn } from '@/lib/utils';

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Users', href: '/users', icon: Users },
  { name: 'Groups', href: '/groups', icon: FolderKanban },
  { name: 'Roles', href: '/roles', icon: Shield },
  { name: 'Permissions', href: '/permissions', icon: Key },
  { name: 'Resources', href: '/resources', icon: FileText },
];

export function Sidebar({ collapsed = false }: { collapsed?: boolean }) {
  const location = useLocation();

  return (
    <div className={`flex h-full ${collapsed ? 'w-20' : 'w-64'} flex-col border-r bg-muted/20`}>
      <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <Shield className="h-6 w-6" />
          {!collapsed && <span className="">Hub</span>}
        </Link>
      </div>
      <div className="flex-1 overflow-auto py-2">
        <nav className={`grid items-start text-sm font-medium ${collapsed ? 'px-1 lg:px-1' : 'px-2 lg:px-4'}`}>
          {navigation.map((item) => {
            const isActive = location.pathname === item.href || (item.href !== '/' && location.pathname.startsWith(item.href));
            return (
              <Link
                key={item.name}
                to={item.href}
                className={cn(
                  'flex items-center rounded-lg py-2 transition-all hover:text-primary',
                  collapsed ? 'gap-0 px-2 justify-center' : 'gap-3 px-3',
                  isActive ? 'bg-muted text-primary' : 'text-muted-foreground'
                )}
              >
                <item.icon className="h-4 w-4" />
                {!collapsed && item.name}
              </Link>
            );
          })}
        </nav>
      </div>
    </div>
  );
}
