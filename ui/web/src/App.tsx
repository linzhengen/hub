import { lazy, Suspense } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router";
import { App as AntdApp, ConfigProvider, Spin, theme as antTheme } from "antd";
import AppLayout from "@/layout/AppLayout";
import { AuthProvider } from "@/providers/AuthProvider";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { AuthErrorHandler } from "@/components/auth/AuthErrorHandler";
import { useTheme } from "@/context/theme";

// ページは遅延読み込みする。まとめて import すると全画面のコードが初回表示で
// ダウンロードされ、チャートライブラリを含む Dashboard まで常に含まれてしまう。
const Dashboard = lazy(() => import("@/pages/Dashboard").then((m) => ({ default: m.Dashboard })));
const Users = lazy(() => import("@/pages/Users").then((m) => ({ default: m.Users })));
const Groups = lazy(() => import("@/pages/system/Groups.tsx").then((m) => ({ default: m.Groups })));
const Roles = lazy(() => import("@/pages/system/Roles.tsx").then((m) => ({ default: m.Roles })));
const Menus = lazy(() => import("@/pages/system/Menus.tsx").then((m) => ({ default: m.Menus })));
const My = lazy(() => import("@/pages/My.tsx").then((m) => ({ default: m.My })));
const Chat = lazy(() => import("@/pages/Chat.tsx").then((m) => ({ default: m.Chat })));
const AccessRequests = lazy(() =>
  import("@/pages/AccessRequests.tsx").then((m) => ({ default: m.AccessRequests })),
);

const PageFallback = () => (
  <div className="flex justify-center py-12">
    <Spin size="large" />
  </div>
);

export default function App() {
  const { theme } = useTheme();

  return (
    <ConfigProvider
      theme={{
        algorithm: theme === "dark" ? antTheme.darkAlgorithm : antTheme.defaultAlgorithm,
      }}
    >
      <AntdApp>
        <AuthProvider>
          <AuthErrorHandler />
          <Router>
            <Suspense fallback={<PageFallback />}>
              <Routes>
                {/* Dashboard Layout */}
                <Route element={
                  <ProtectedRoute>
                    <AppLayout />
                  </ProtectedRoute>
                }>
                  <Route index path="/" element={<Dashboard />} />
                  <Route path="/dashboard" element={<Dashboard />} />
                  <Route path="/users" element={<Users />} />
                  <Route path="/system/groups" element={<Groups />} />
                  <Route path="/system/roles" element={<Roles />} />
                  <Route path="/system/menus" element={<Menus />} />
                  <Route path="/my" element={<My />} />
                  <Route path="/chat" element={<Chat />} />
                  <Route path="/access-requests" element={<AccessRequests />} />
                </Route>

                {/* Fallback Route */}
                <Route path="*" element={<div>Page Not Found</div>} />
              </Routes>
            </Suspense>
          </Router>
        </AuthProvider>
      </AntdApp>
    </ConfigProvider>
  );
}
