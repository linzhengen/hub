import React from 'react';
import { CustomLayout } from './CustomLayout';

// カスタムレイアウトコンポーネント
// ThemedLayoutの代わりにCustomLayoutを使用
export const RefineLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return <CustomLayout>{children}</CustomLayout>;
};
