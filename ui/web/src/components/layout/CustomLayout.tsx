import React, { useState } from 'react';
import { Layout, Button, theme, ConfigProvider, Dropdown, Avatar, Space, Badge } from 'antd';
import { MenuUnfoldOutlined, MenuFoldOutlined, UserOutlined, LogoutOutlined, BellOutlined, SettingOutlined, SearchOutlined } from '@ant-design/icons';
import { CustomSider } from './CustomSider';
import { useLogout, useGetIdentity } from '@refinedev/core';

const { Header, Content, Footer } = Layout;

interface CustomLayoutProps {
  children: React.ReactNode;
}

export const CustomLayout: React.FC<CustomLayoutProps> = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const { token } = theme.useToken();
  const { mutate: logout } = useLogout();
  const { data: identity } = useGetIdentity();

  const handleLogout = () => {
    console.log('CustomLayout: Logout menu item clicked');
    logout();
  };

  const dropdownItems = [
    {
      key: 'logout',
      label: 'ログアウト',
      icon: <LogoutOutlined />,
      onClick: handleLogout,
    },
  ];

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#3b82f6',
          borderRadius: 8,
          colorBgLayout: '#f8fafc',
          colorText: '#1e293b',
          colorTextSecondary: '#64748b',
          colorBorder: '#e2e8f0',
          colorBgContainer: '#ffffff',
          fontSize: 14,
          wireframe: false,
        },
        components: {
          Layout: {
            bodyBg: '#f8fafc',
            headerBg: '#ffffff',
            headerHeight: 64,
            headerPadding: '0 24px',
            siderBg: '#ffffff',
            triggerBg: '#f1f5f9',
            triggerColor: '#64748b',
          },
          Menu: {
            itemBg: 'transparent',
            itemHoverBg: '#f1f5f9',
            itemSelectedBg: '#eff6ff',
            itemSelectedColor: '#3b82f6',
            itemMarginInline: 0,
            itemPaddingInline: 16,
            itemHeight: 40,
            itemBorderRadius: 6,
            iconSize: 18,
            collapsedIconSize: 18,
            subMenuItemBg: 'transparent',
            darkItemBg: '#1e293b',
            darkItemHoverBg: '#334155',
            darkItemSelectedBg: '#1e40af',
            darkItemSelectedColor: '#60a5fa',
            horizontalItemBorderRadius: 6,
            horizontalItemSelectedColor: '#3b82f6',
            horizontalItemSelectedBg: '#eff6ff',
            horizontalItemHoverBg: '#f1f5f9',
          },
          Card: {
            paddingLG: 24,
            paddingMD: 20,
            paddingSM: 16,
            borderRadiusLG: 12,
            borderRadius: 10,
            borderRadiusSM: 8,
            boxShadowTertiary: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
          },
          Button: {
            borderRadius: 8,
            borderRadiusSM: 6,
            borderRadiusLG: 10,
            controlHeight: 36,
            controlHeightSM: 32,
            controlHeightLG: 40,
          },
          Input: {
            borderRadius: 8,
            controlHeight: 36,
            controlHeightSM: 32,
            controlHeightLG: 40,
          },
          Table: {
            borderRadius: 8,
            headerBg: '#f8fafc',
            headerSplitColor: 'transparent',
            rowHoverBg: '#f1f5f9',
          },
        },
      }}
    >
      <Layout style={{ minHeight: '100vh' }}>
        <CustomSider collapsed={collapsed} />
        <Layout style={{ marginLeft: collapsed ? 80 : 256, transition: 'all 0.2s' }}>
          <Header
            style={{
              padding: '0 24px',
              backgroundColor: token.colorBgContainer,
              borderBottom: `1px solid ${token.colorBorder}`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              height: '64px',
            }}
          >
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{ fontSize: '16px' }}
            />

            {/* 検索バー */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '16px', flex: 1, marginLeft: '24px' }}>
              <div style={{ position: 'relative', width: '280px' }}>
                <SearchOutlined style={{
                  position: 'absolute',
                  left: '12px',
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: '#94a3b8',
                  fontSize: '16px'
                }} />
                <input
                  type="text"
                  placeholder="Search..."
                  style={{
                    width: '100%',
                    padding: '8px 12px 8px 40px',
                    borderRadius: '8px',
                    border: '1px solid #e2e8f0',
                    backgroundColor: '#f8fafc',
                    fontSize: '14px',
                    outline: 'none',
                    transition: 'all 0.2s'
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = '#3b82f6';
                    e.target.style.backgroundColor = '#ffffff';
                    e.target.style.boxShadow = '0 0 0 3px rgba(59, 130, 246, 0.1)';
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = '#e2e8f0';
                    e.target.style.backgroundColor = '#f8fafc';
                    e.target.style.boxShadow = 'none';
                  }}
                />
              </div>
            </div>

            {/* 右側: ユーザー情報とアイコン */}
            <Space size="middle">
              {/* 通知アイコン */}
              <Badge count={3} size="small">
                <Button
                  type="text"
                  icon={<BellOutlined style={{ fontSize: '18px', color: '#64748b' }} />}
                  style={{ padding: '8px', borderRadius: '8px' }}
                />
              </Badge>

              {/* 設定アイコン */}
              <Button
                type="text"
                icon={<SettingOutlined style={{ fontSize: '18px', color: '#64748b' }} />}
                style={{ padding: '8px', borderRadius: '8px' }}
              />

              {/* ユーザープロフィール */}
              {identity && identity.id ? (
                <Dropdown menu={{ items: dropdownItems }} placement="bottomRight" trigger={['click']}>
                  <Space style={{ cursor: 'pointer', padding: '4px 8px', borderRadius: '8px', transition: 'background-color 0.2s' }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = '#f1f5f9';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                    }}
                  >
                    <Avatar
                      size="default"
                      style={{
                        backgroundColor: '#3b82f6',
                        color: '#ffffff'
                      }}
                      icon={identity.avatar ? <img src={identity.avatar} alt={identity.name} /> : <UserOutlined />}
                    />
                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
                      <span style={{ fontWeight: '500', fontSize: '14px', color: '#1e293b' }}>{identity.name || 'User'}</span>
                      <span style={{ fontSize: '12px', color: '#64748b' }}>{identity.email || 'Administrator'}</span>
                    </div>
                  </Space>
                </Dropdown>
              ) : (
                <div style={{ padding: '8px', color: '#64748b', fontSize: '14px' }}>
                  認証中...
                </div>
              )}
            </Space>
          </Header>
          <Content
            style={{
              margin: '24px',
              padding: 24,
              minHeight: 280,
              backgroundColor: token.colorBgContainer,
              borderRadius: token.borderRadiusLG,
              overflow: 'auto',
            }}
          >
            {children}
          </Content>
          <Footer
            style={{
              textAlign: 'center',
              padding: '20px 24px',
              color: '#64748b',
              fontSize: '14px',
              borderTop: '1px solid #e2e8f0',
              backgroundColor: '#ffffff'
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span>© 2026 AI Hub System. All rights reserved.</span>
              <div style={{ display: 'flex', gap: '16px' }}>
                <a href="#" style={{ color: '#64748b', textDecoration: 'none' }}>Privacy Policy</a>
                <a href="#" style={{ color: '#64748b', textDecoration: 'none' }}>Terms of Service</a>
                <a href="#" style={{ color: '#64748b', textDecoration: 'none' }}>Help Center</a>
              </div>
            </div>
          </Footer>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
};
