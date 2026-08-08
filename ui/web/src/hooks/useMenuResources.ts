import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/providers/auth";
import { userService } from "@/services/user";
import type { Menu } from "@/services/user";

export const MENU_RESOURCES_QUERY_KEY = ["meMenus"];

/**
 * ログインユーザーのロールに基づいてフィルタリングされたメニューツリーを取得する。
 * サーバーサイドで RBAC 評価済みのため、返却されたメニューはすべて表示可能。
 */
export const useMenuResources = () => {
  const { isAuthenticated } = useAuth();

  return useQuery<Menu[]>({
    queryKey: MENU_RESOURCES_QUERY_KEY,
    queryFn: async () => {
      const res = await userService.getMeMenus();
      return res.menus ?? [];
    },
    enabled: isAuthenticated,
  });
};
