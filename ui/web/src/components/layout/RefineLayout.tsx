import React from 'react';
import { ThemedLayout, ThemedTitle } from '@refinedev/antd';
import { useLogout, useGetIdentity } from '@refinedev/core';
import { ConfigProvider, Button, Dropdown, Avatar, Space } from 'antd';
import { UserOutlined, LogoutOutlined } from '@ant-design/icons';

// カスタムレイアウトコンポーネント
// 必要に応じてThemedLayoutV2をカスタマイズするために使用
export const RefineLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1890ff',
          borderRadius: 6,
          colorBgLayout: '#f5f5f5',
        },
        components: {
          Layout: {
            bodyBg: '#ffffff',
            headerBg: '#ffffff',
            siderBg: '#ffffff',
          },
          Menu: {
            itemBg: 'transparent',
            itemHoverBg: '#f0f0f0',
            itemSelectedBg: '#e6f7ff',
            itemSelectedColor: '#1890ff',
          },
        },
      }}
    >
      <ThemedLayout
        Title={({ collapsed }) => (
          <ThemedTitle
            collapsed={collapsed}
            text="AI Hub"
            icon={null}
          />
        )}
        // ヘッダーのカスタマイズ
        Header={() => {
          const { mutate: logout } = useLogout();
          const { data: identity } = useGetIdentity();

          const handleLogout = () => {
            console.log('RefineLayout: Logout button clicked');
            logout();
          };

          const items = [
            {
              key: 'logout',
              label: 'ログアウト',
              icon: <LogoutOutlined />,
              onClick: handleLogout,
            },
          ];

          return (
            <div style={{ padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'flex-end', height: '64px' }}>
              <Space>
                {identity && (
                  <Dropdown menu={{ items }} placement="bottomRight">
                    <Space style={{ cursor: 'pointer' }}>
                      <Avatar
                        size="small"
                        icon={identity.avatar ? <img src={identity.avatar} alt={identity.name} /> : <UserOutlined />}
                      />
                      <span>{identity.name}</span>
                    </Space>
                  </Dropdown>
                )}
                {!identity && (
                  <Button type="primary" onClick={handleLogout}>
                    ログイン
                  </Button>
                )}
              </Space>
            </div>
          );
        }}
        // フッターのカスタマイズ
        Footer={() => (
          <div style={{ textAlign: 'center', padding: '16px', color: '#999' }}>
            AI Hub System © 2026
          </div>
        )}
        // オンボーディングツアーの無効化
        // その他のプロパティ
      >
        {children}
      </ThemedLayout>
    </ConfigProvider>
  );
};

// シンプルなレイアウト（サイドバーなし）
export const SimpleLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1890ff',
          borderRadius: 6,
        },
      }}
    >
      <div style={{ minHeight: '100vh', backgroundColor: '#f5f5f5' }}>
        <main style={{ padding: '24px' }}>
          {children}
        </main>
      </div>
    </ConfigProvider>
  );
};
