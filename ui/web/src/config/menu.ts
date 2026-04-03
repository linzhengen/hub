import React from 'react';
import {
  UserOutlined,
  TeamOutlined,
  SafetyCertificateOutlined,
  KeyOutlined,
  AppstoreOutlined,
  SettingOutlined,
  DashboardOutlined
} from '@ant-design/icons';
import { ResourceProps } from '@refinedev/core';

// menus.yamlの構造をTypeScript形式に変換
export interface MenuItem {
  name: string;
  path: string;
  icon?: React.ReactNode;
  hideInMenu?: boolean;
  children?: MenuItem[];
}

export const menuItems: MenuItem[] = [
  {
    name: 'マイページ',
    path: '/my',
    icon: React.createElement(UserOutlined),
    hideInMenu: true, // metadata.hideInMenu: "true" に対応
  },
  {
    name: 'ダッシュボード',
    path: '/',
    icon: React.createElement(DashboardOutlined),
  },
  {
    name: 'ユーザー管理',
    path: '/user',
    icon: React.createElement(UserOutlined),
  },
  {
    name: 'システム管理',
    path: '/system',
    icon: React.createElement(SettingOutlined),
    children: [
      {
        name: 'ロール管理',
        path: '/system/roles',
        icon: React.createElement(SafetyCertificateOutlined),
      },
      {
        name: 'グループ管理',
        path: '/system/groups',
        icon: React.createElement(TeamOutlined),
      },
      {
        name: 'メニュー管理',
        path: '/system/menus',
        icon: React.createElement(AppstoreOutlined),
      },
    ],
  },
];

// Refineのresources形式に変換
export const resources: ResourceProps[] = [
  {
    name: 'dashboard',
    list: '/',
    meta: {
      label: 'ダッシュボード',
      icon: React.createElement(DashboardOutlined),
    },
  },
  {
    name: 'user',
    list: '/users',
    meta: {
      label: 'ユーザー管理',
      icon: React.createElement(UserOutlined),
    },
  },
  {
    name: 'group',
    list: '/system/group',
    meta: {
      label: 'グループ管理',
      icon: React.createElement(TeamOutlined),
    },
  },
  {
    name: 'role',
    list: '/system/role',
    meta: {
      label: 'ロール管理',
      icon: React.createElement(SafetyCertificateOutlined),
    },
  },
  {
    name: 'permission',
    list: '/system/permission',
    meta: {
      label: '権限管理',
      icon: React.createElement(KeyOutlined),
    },
  },
  {
    name: 'resource',
    list: '/system/resource',
    meta: {
      label: 'リソース管理',
      icon: React.createElement(AppstoreOutlined),
    },
  },
];

// メニュー項目をresourcesから生成するヘルパー関数
export function getMenuItemsFromResources(): MenuItem[] {
  return resources.map(resource => ({
    name: resource.meta?.label || resource.name,
    path: typeof resource.list === 'string' ? resource.list : '/',
    icon: resource.meta?.icon,
  }));
}
