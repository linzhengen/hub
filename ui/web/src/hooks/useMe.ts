import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/providers/AuthProvider";
import { userService } from "@/services/user";
import type { GetMeResponse } from "@/services/user";

export const ME_QUERY_KEY = ["me"];

/**
 * ログイン中ユーザーのプロフィールを取得する。
 *
 * `useAuth()` の `user` は Keycloak のアクセストークン由来で、プロフィールを
 * 更新してもトークンが再発行されるまで古い値のままになる。表示名やメール
 * アドレスはこのフックを参照すること（トークンにしか無い `emailVerified` は
 * 引き続き `useAuth()` から取る）。
 */
export const useMe = () => {
  const { isAuthenticated } = useAuth();

  return useQuery<GetMeResponse>({
    queryKey: ME_QUERY_KEY,
    queryFn: () => userService.getMe(),
    enabled: isAuthenticated,
  });
};
