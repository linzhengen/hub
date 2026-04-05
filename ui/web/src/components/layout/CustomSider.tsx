import React from 'react';
import { useMenu, useGetIdentity } from '@refinedev/core';
import { Layout, Menu, Avatar, Space, Typography, theme } from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  SafetyCertificateOutlined,
  KeyOutlined,
  AppstoreOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import { Link } from 'react-router-dom';

const { Sider } = Layout;
const { Text } = Typography;

interface CustomSiderProps {
  collapsed: boolean;
}

export const CustomSider: React.FC<CustomSiderProps> = ({ collapsed }) => {
  const { token } = theme.useToken();
  const { data: identity } = useGetIdentity();
  const { menuItems, selectedKey } = useMenu();

  // Map menu items to antd Menu items format
  const mapMenuItems = (items: any[]): any[] => {
    return items.map((item) => {
      const menuItem: any = {
        key: item.key || item.name,
        label: item.label || item.name,
        icon: item.icon,
      };

      if (item.route) {
        menuItem.label = <Link to={item.route}>{menuItem.label}</Link>;
      }

      if (item.children && item.children.length > 0) {
        menuItem.children = mapMenuItems(item.children);
      }

      return menuItem;
    });
  };

  // Default menu items if useMenu doesn't return anything
  const defaultMenuItems = [
    {
      key: 'dashboard',
      label: <Link to="/">ダッシュボード</Link>,
      icon: <DashboardOutlined />,
    },
    {
      key: 'user',
      label: <Link to="/users">ユーザー管理</Link>,
      icon: <UserOutlined />,
    },
    {
      key: 'system',
      label: 'システム管理',
      icon: <SettingOutlined />,
      children: [
        {
          key: 'role',
          label: <Link to="/system/role">ロール管理</Link>,
          icon: <SafetyCertificateOutlined />,
        },
        {
          key: 'group',
          label: <Link to="/system/group">グループ管理</Link>,
          icon: <TeamOutlined />,
        },
        {
          key: 'permission',
          label: <Link to="/system/permission">権限管理</Link>,
          icon: <KeyOutlined />,
        },
        {
          key: 'resource',
          label: <Link to="/system/resource">リソース管理</Link>,
          icon: <AppstoreOutlined />,
        },
      ],
    },
  ];

  const items = menuItems && menuItems.length > 0 ? mapMenuItems(menuItems) : defaultMenuItems;


  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={collapsed}
      width={256}
      style={{
        backgroundColor: token.colorBgContainer,
        borderRight: `1px solid ${token.colorBorder}`,
        height: '100vh',
        position: 'fixed',
        left: 0,
        top: 0,
        bottom: 0,
        zIndex: 100,
      }}
    >
      {/* Logo/Title */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          height: '64px',
          padding: collapsed ? '0 16px' : '0 24px',
          borderBottom: `1px solid ${token.colorBorder}`,
        }}
      >
        {!collapsed ? (
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <div
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                backgroundColor: token.colorPrimary,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#ffffff',
                fontWeight: '600',
                fontSize: '16px',
              }}
            >
              H
            </div>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              <Text strong style={{ fontSize: '16px', color: token.colorText }}>
                AI Hub
              </Text>
              <Text type="secondary" style={{ fontSize: '12px' }}>
                Admin Dashboard
              </Text>
            </div>
          </div>
        ) : (
          <div
            style={{
              width: '32px',
              height: '32px',
              borderRadius: '8px',
              backgroundColor: token.colorPrimary,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#ffffff',
              fontWeight: '600',
              fontSize: '16px',
              margin: '0 auto',
            }}
          >
            H
          </div>
        )}
      </div>

      {/* Menu */}
      <div style={{ padding: '16px 0', flex: 1, overflow: 'auto' }}>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey || 'dashboard']}
          items={items}
          style={{
            border: 'none',
            backgroundColor: 'transparent',
          }}
        />
      </div>

      {/* User Profile */}
      <div
        style={{
          padding: '16px',
          borderTop: `1px solid ${token.colorBorder}`,
          backgroundColor: token.colorBgLayout,
        }}
      >
        {identity && identity.id ? (
          <Space direction="vertical" size="small" style={{ width: '100%' }}>
            <Space>
              <Avatar
                size="default"
                style={{
                  backgroundColor: token.colorPrimary,
                  color: '#ffffff',
                }}
                icon={identity.avatar ? <img src={identity.avatar} alt={identity.name} /> : <UserOutlined />}
              />
              {!collapsed && (
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
                  <Text strong style={{ fontSize: '14px', color: token.colorText }}>
                    {identity.name || 'User'}
                  </Text>
                  <Text type="secondary" style={{ fontSize: '12px' }}>
                    {identity.email || 'Administrator'}
                  </Text>
                </div>
              )}
            </Space>
          </Space>
        ) : (
          <div style={{ padding: '8px', color: token.colorTextSecondary, fontSize: '14px', textAlign: 'center' }}>
            認証中...
          </div>
        )}
      </div>
    </Sider>
  );
};
