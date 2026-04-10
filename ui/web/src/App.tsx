import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { ConfigProvider, theme as antTheme } from "antd";
import AppLayout from "@/layout/AppLayout";
import { Dashboard } from "@/pages/Dashboard";
import { Users } from "@/pages/Users";
import { Groups } from "@/pages/system/Groups.tsx";
import { Roles } from "@/pages/system/Roles.tsx";
import { Resources } from "@/pages/system/Resources.tsx";
import { Permissions } from "@/pages/system/Permissions.tsx";
import { Profile } from "@/pages/Profile";
import { AuthProvider } from "@/providers/AuthProvider";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { AuthErrorHandler } from "@/components/auth/AuthErrorHandler";
import { useTheme } from "@/context/ThemeContext";

export default function App() {
  const { theme } = useTheme();

  return (
    <ConfigProvider
      theme={{
        algorithm: theme === "dark" ? antTheme.darkAlgorithm : antTheme.defaultAlgorithm,
      }}
    >
      <AuthProvider>
        <AuthErrorHandler />
        <Router>
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
              <Route path="/system/resources" element={<Resources />} />
              <Route path="/system/permissions" element={<Permissions />} />
              <Route path="/profile" element={<Profile />} />
            </Route>

            {/* Fallback Route */}
            <Route path="*" element={<div>Page Not Found</div>} />
          </Routes>
        </Router>
      </AuthProvider>
    </ConfigProvider>
  );
}
