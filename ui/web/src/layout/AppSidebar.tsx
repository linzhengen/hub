import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link, useLocation } from "react-router";

import {
  ChevronDownIcon,
  HorizontaLDots,
  FolderIcon,
  GridIcon,
  UserIcon,
  GroupIcon,
  LockIcon,
  BoltIcon,
  BoxIcon,
  CalenderIcon,
  ChatIcon,
  DocsIcon,
  EnvelopeIcon,
  FileIcon,
  ListIcon,
  MailIcon,
  PageIcon,
  PieChartIcon,
  PlugInIcon,
  ShootingStarIcon,
  TableIcon,
  TaskIcon,
  TimeIcon,
  UserCircleIcon,
} from "@/icons";
import { Key } from "lucide-react";
import { useSidebar } from "@/context/sidebar";
import { useMenuResources } from "@/hooks/useMenuResources";
import type { Menu } from "@/services/user";

type NavItem = {
  name: string;
  icon: React.ReactNode;
  path?: string;
  subItems?: { name: string; icon?: React.ReactNode; path: string; pro?: boolean; new?: boolean }[];
};

/** metadata.icon の文字列から React ノードを返す。未知のキーは FolderIcon にフォールバック。 */
const resolveMenuIcon = (iconName: string | undefined): React.ReactNode => {
  // 大文字小文字・プレフィックス（"mdi:"等）を正規化して比較する
  const normalized = iconName?.replace(/^[^:]+:/, "").toLowerCase() ?? "";
  switch (normalized) {
    case "grid":              return <GridIcon />;
    case "gridicon":          return <GridIcon />;
    case "user":              return <UserIcon />;
    case "usericon":          return <UserIcon />;
    case "user-circle":       return <UserCircleIcon />;
    case "account-circle":    return <UserCircleIcon />;
    case "usercircleicon":    return <UserCircleIcon />;
    case "group":             return <GroupIcon />;
    case "groupicon":         return <GroupIcon />;
    case "lock":              return <LockIcon />;
    case "lockicon":          return <LockIcon />;
    case "settings":          return <LockIcon />;
    case "folder":            return <FolderIcon />;
    case "foldericon":        return <FolderIcon />;
    case "key":               return <Key />;
    case "bolt":              return <BoltIcon />;
    case "bolticon":          return <BoltIcon />;
    case "box":               return <BoxIcon />;
    case "boxicon":           return <BoxIcon />;
    case "calendar":          return <CalenderIcon />;
    case "calendericon":      return <CalenderIcon />;
    case "chat":              return <ChatIcon />;
    case "chaticon":          return <ChatIcon />;
    case "docs":              return <DocsIcon />;
    case "docsicon":          return <DocsIcon />;
    case "envelope":          return <EnvelopeIcon />;
    case "envelopeicon":      return <EnvelopeIcon />;
    case "file":              return <FileIcon />;
    case "fileicon":          return <FileIcon />;
    case "list":              return <ListIcon />;
    case "listicon":          return <ListIcon />;
    case "mail":              return <MailIcon />;
    case "mailicon":          return <MailIcon />;
    case "page":              return <PageIcon />;
    case "pageicon":          return <PageIcon />;
    case "pie-chart":         return <PieChartIcon />;
    case "piecharticon":      return <PieChartIcon />;
    case "plug-in":           return <PlugInIcon />;
    case "pluginicon":        return <PlugInIcon />;
    case "star":              return <ShootingStarIcon />;
    case "shootingstaricon":  return <ShootingStarIcon />;
    case "table":             return <TableIcon />;
    case "tableicon":         return <TableIcon />;
    case "task":              return <TaskIcon />;
    case "taskicon":          return <TaskIcon />;
    case "time":              return <TimeIcon />;
    case "timeicon":          return <TimeIcon />;
    default:                  return <FolderIcon />;
  }
};

/**
 * サーバーから返却された Menu ツリーを NavItem[] に変換する。
 * hideInMenu な項目はサーバーが除外済みだが、フロントでも念のためスキップする。
 * meta.order 昇順はサーバーがソート済みのためここでは行わない。
 */
const buildNavItems = (menus: Menu[]): NavItem[] =>
  menus
    .filter((m) => !m.meta?.hideInMenu)
    .map((m) => {
      const children = (m.children ?? []).filter((c) => !c.meta?.hideInMenu);
      const navItem: NavItem = {
        name: m.meta?.title ?? m.name ?? "",
        icon: resolveMenuIcon(m.meta?.icon),
        path: children.length === 0 ? (m.path ?? undefined) : undefined,
      };
      if (children.length > 0) {
        navItem.subItems = children.map((child) => ({
          name: child.meta?.title ?? child.name ?? "",
          icon: resolveMenuIcon(child.meta?.icon),
          path: child.path ?? "",
        }));
      }
      return navItem;
    });

const othersItems: NavItem[] = [];

type Submenu = { type: "main" | "others"; index: number };

/** 手動で開閉したサブメニューと、その操作を行ったルート */
type ManualSubmenu = { pathname: string; submenu: Submenu | null };

/** 現在のパスを含むサブメニューを返す。該当しなければ null */
const findSubmenuForPath = (pathname: string, mainItems: NavItem[]): Submenu | null => {
  const groups = [
    { type: "main", items: mainItems },
    { type: "others", items: othersItems },
  ] as const;

  let match: Submenu | null = null;
  for (const group of groups) {
    group.items.forEach((nav, index) => {
      if (nav.subItems?.some((subItem) => subItem.path === pathname)) {
        match = { type: group.type, index };
      }
    });
  }
  return match;
};

const AppSidebar: React.FC = () => {
  const { isExpanded, isMobileOpen, isHovered, setIsHovered } = useSidebar();
  const location = useLocation();
  const { data: menuResources, isLoading: isMenuLoading } = useMenuResources();

  const dynamicNavItems = useMemo(
    () => (menuResources ? buildNavItems(menuResources) : null),
    [menuResources],
  );

  // API取得中または取得失敗時は静的フォールバック
  const activeNavItems: NavItem[] = useMemo(
    () =>
      dynamicNavItems ??
      [
        { icon: <GridIcon />, name: "Dashboard", path: "/dashboard" },
        { icon: <UserIcon />, name: "Users", path: "/users" },
        {
          name: "Systems",
          icon: <LockIcon />,
          subItems: [
            { name: "Groups", icon: <GroupIcon />, path: "/system/groups" },
            { name: "Roles", icon: <Key />, path: "/system/roles" },
            { name: "Menus", icon: <FolderIcon />, path: "/system/menus" },
          ],
        },
      ],
    [dynamicNavItems],
  );

  const [subMenuHeight, setSubMenuHeight] = useState<Record<string, number>>(
    {}
  );
  const subMenuRefs = useRef<Record<string, HTMLDivElement | null>>({});

  // const isActive = (path: string) => location.pathname === path;
  const isActive = useCallback(
    (path: string) => location.pathname === path,
    [location.pathname]
  );

  // どのサブメニューを開くかは現在のルートから決まる純粋な導出なので、effect で
  // setState せず描画時に計算する。手動トグルはそのルートにいる間だけ上書きし、
  // 遷移すると破棄される（effect がナビゲーションのたびに上書きしていた従来の
  // 挙動と同じ）。
  const routeSubmenu = useMemo(
    () => findSubmenuForPath(location.pathname, activeNavItems),
    [location.pathname, activeNavItems],
  );
  const [manualSubmenu, setManualSubmenu] = useState<ManualSubmenu | null>(null);
  const openSubmenu = useMemo(
    () => (manualSubmenu?.pathname === location.pathname ? manualSubmenu.submenu : routeSubmenu),
    [manualSubmenu, location.pathname, routeSubmenu]
  );

  useEffect(() => {
    if (openSubmenu !== null) {
      const key = `${openSubmenu.type}-${openSubmenu.index}`;
      if (subMenuRefs.current[key]) {
        setSubMenuHeight((prevHeights) => ({
          ...prevHeights,
          [key]: subMenuRefs.current[key]?.scrollHeight || 0,
        }));
      }
    }
  }, [openSubmenu]);

  const handleSubmenuToggle = (index: number, menuType: "main" | "others") => {
    const isAlreadyOpen = openSubmenu?.type === menuType && openSubmenu?.index === index;
    setManualSubmenu({
      pathname: location.pathname,
      submenu: isAlreadyOpen ? null : { type: menuType, index },
    });
  };

  const renderMenuItems = (items: NavItem[], menuType: "main" | "others") => (
    <ul className="flex flex-col gap-4">
      {items.map((nav, index) => (
        <li key={nav.name}>
          {nav.subItems ? (
            <button
              onClick={() => handleSubmenuToggle(index, menuType)}
              className={`menu-item group ${
                openSubmenu?.type === menuType && openSubmenu?.index === index
                  ? "menu-item-active"
                  : "menu-item-inactive"
              } cursor-pointer ${
                !isExpanded && !isHovered
                  ? "lg:justify-center"
                  : "lg:justify-start"
              }`}
            >
              <span
                className={`menu-item-icon-size  ${
                  openSubmenu?.type === menuType && openSubmenu?.index === index
                    ? "menu-item-icon-active"
                    : "menu-item-icon-inactive"
                }`}
              >
                {nav.icon}
              </span>
              {(isExpanded || isHovered || isMobileOpen) && (
                <span className="menu-item-text">{nav.name}</span>
              )}
              {(isExpanded || isHovered || isMobileOpen) && (
                <ChevronDownIcon
                  className={`ml-auto w-5 h-5 transition-transform duration-200 ${
                    openSubmenu?.type === menuType &&
                    openSubmenu?.index === index
                      ? "rotate-180 text-brand-500"
                      : ""
                  }`}
                />
              )}
            </button>
          ) : (
            nav.path && (
              <Link
                to={nav.path}
                className={`menu-item group ${
                  isActive(nav.path) ? "menu-item-active" : "menu-item-inactive"
                }`}
              >
                <span
                  className={`menu-item-icon-size ${
                    isActive(nav.path)
                      ? "menu-item-icon-active"
                      : "menu-item-icon-inactive"
                  }`}
                >
                  {nav.icon}
                </span>
                {(isExpanded || isHovered || isMobileOpen) && (
                  <span className="menu-item-text">{nav.name}</span>
                )}
              </Link>
            )
          )}
          {nav.subItems && (isExpanded || isHovered || isMobileOpen) && (
            <div
              ref={(el) => {
                subMenuRefs.current[`${menuType}-${index}`] = el;
              }}
              className="overflow-hidden transition-all duration-300"
              style={{
                height:
                  openSubmenu?.type === menuType && openSubmenu?.index === index
                    ? `${subMenuHeight[`${menuType}-${index}`]}px`
                    : "0px",
              }}
            >
              <ul className="mt-2 space-y-1 ml-9">
                {nav.subItems.map((subItem) => (
                  <li key={subItem.name}>
                    <Link
                      to={subItem.path}
                      className={`menu-dropdown-item flex items-center gap-3 ${
                        isActive(subItem.path)
                          ? "menu-dropdown-item-active"
                          : "menu-dropdown-item-inactive"
                      }`}
                    >
                      {subItem.icon && (
                        <span className="flex-shrink-0">{subItem.icon}</span>
                      )}
                      {subItem.name}
                      <span className="flex items-center gap-1 ml-auto">
                        {subItem.new && (
                          <span
                            className={`ml-auto ${
                              isActive(subItem.path)
                                ? "menu-dropdown-badge-active"
                                : "menu-dropdown-badge-inactive"
                            } menu-dropdown-badge`}
                          >
                            new
                          </span>
                        )}
                        {subItem.pro && (
                          <span
                            className={`ml-auto ${
                              isActive(subItem.path)
                                ? "menu-dropdown-badge-active"
                                : "menu-dropdown-badge-inactive"
                            } menu-dropdown-badge`}
                          >
                            pro
                          </span>
                        )}
                      </span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </li>
      ))}
    </ul>
  );

  return (
    <aside
      className={`fixed mt-16 flex flex-col lg:mt-0 top-0 px-5 left-0 bg-white dark:bg-gray-900 dark:border-gray-800 text-gray-900 h-screen transition-all duration-300 ease-in-out z-50 border-r border-gray-200
        ${
          isExpanded || isMobileOpen
            ? "w-[260px]"
            : isHovered
            ? "w-[260px]"
            : "w-[90px]"
        }
        ${isMobileOpen ? "translate-x-0" : "-translate-x-full"}
        lg:translate-x-0 transition-colors duration-300`}
      onMouseEnter={() => !isExpanded && setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <div
        className={`py-8 flex ${
          !isExpanded && !isHovered ? "lg:justify-center" : "justify-start"
        }`}
      >
        <Link to="/">
          {isExpanded || isHovered || isMobileOpen ? (
            <>
              <img
                className="dark:hidden"
                src="/images/logo/logo.svg"
                alt="Logo"
                width={150}
                height={40}
              />
              <img
                className="hidden dark:block"
                src="/images/logo/logo-dark.svg"
                alt="Logo"
                width={150}
                height={40}
              />
            </>
          ) : (
            <img
              src="/images/logo/logo-icon.svg"
              alt="Logo"
              width={32}
              height={32}
            />
          )}
        </Link>
      </div>
      <div className="flex flex-col overflow-y-auto duration-300 ease-linear no-scrollbar">
        <nav className="mb-6">
          <div className="flex flex-col gap-4">
            <div>
              <h2
                className={`mb-4 text-xs uppercase flex leading-[20px] text-gray-400 ${
                  !isExpanded && !isHovered
                    ? "lg:justify-center"
                    : "justify-start"
                }`}
              >
                {isExpanded || isHovered || isMobileOpen ? (
                  "Menu"
                ) : (
                  <HorizontaLDots className="size-6" />
                )}
              </h2>
              {isMenuLoading && !dynamicNavItems ? (
                <div className="flex flex-col gap-3 px-1">
                  {[1, 2, 3].map((i) => (
                    <div
                      key={i}
                      className="h-9 rounded-lg bg-gray-100 dark:bg-gray-800 animate-pulse"
                    />
                  ))}
                </div>
              ) : (
                renderMenuItems(activeNavItems, "main")
              )}
            </div>
          </div>
        </nav>
      </div>
    </aside>
  );
};

export default AppSidebar;
