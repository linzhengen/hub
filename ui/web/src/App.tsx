/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { Refine } from '@refinedev/core';
import { RefineLayout } from '@/components/layout/RefineLayout';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from '@/components/ui/sonner';
import { Dashboard } from '@/pages/Dashboard';
import { Users } from '@/pages/Users';
import { Groups } from '@/pages/Groups';
import { Roles } from '@/pages/Roles';
import { Permissions } from '@/pages/Permissions';
import { Resources } from '@/pages/Resources';
import { authProvider } from '@/providers/authProvider';
import { dataProvider } from '@/providers/dataProvider';
import { resources } from '@/config/menu';

const queryClient = new QueryClient();

function AppContent() {
  return (
    <>
      <RefineLayout>
        <Routes>
          <Route index element={<Dashboard />} />
          <Route path="users" element={<Users />} />
          <Route path="system/group" element={<Groups />} />
          <Route path="system/role" element={<Roles />} />
          <Route path="system/permission" element={<Permissions />} />
          <Route path="system/resource" element={<Resources />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </RefineLayout>
      <Toaster />
    </>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Refine
          dataProvider={dataProvider}
          authProvider={authProvider}
          resources={resources}
          options={{
            syncWithLocation: true,
            warnWhenUnsavedChanges: true,
            disableTelemetry: true,
          }}
        >
          <AppContent />
        </Refine>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
