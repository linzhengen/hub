import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { ConfigProvider, theme as antTheme } from "antd";
import AppLayout from "@/layout/AppLayout";
import { Dashboard } from "@/pages/Dashboard";
import { Users } from "@/pages/Users";
import { Groups } from "@/pages/Groups";
import { Roles } from "@/pages/Roles";
import { Resources } from "@/pages/Resources";
import { Permissions } from "@/pages/Permissions";
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
              <Route path="/groups" element={<Groups />} />
              <Route path="/roles" element={<Roles />} />
              <Route path="/resources" element={<Resources />} />
              <Route path="/permissions" element={<Permissions />} />
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
