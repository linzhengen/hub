import React, { useState, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/providers/AuthProvider';
import { ME_QUERY_KEY, useMe } from '@/hooks/useMe';
import { Card, Descriptions, Tag, Button, Form, Input, Modal, Divider, Spin } from 'antd';
import { UserOutlined, MailOutlined, SafetyOutlined, EditOutlined, SaveOutlined } from '@ant-design/icons';
import { toast } from 'sonner';
import { userService } from '@/services/user';
import type { GetMeResponse, UpdateMeRequest } from '@/services/user';

interface ProfileFormValues {
  name: string;
  email: string;
}

export function My() {
  const { user: authUser } = useAuth();
  const queryClient = useQueryClient();
  const [isEditing, setIsEditing] = useState(false);
  const [editForm] = Form.useForm();

  const { data: meResponse, isLoading, error } = useMe();
  const apiUser = meResponse?.user;
  const apiGroups = meResponse?.groups;

  // 取得に失敗しても authUser にフォールバックして表示は続けるため、通知のみ行う
  useEffect(() => {
    if (!error) return;
    console.error('Failed to fetch user from API:', error);
    toast.error('Failed to load profile data from server');
  }, [error]);

  // 表示用のユーザー情報をマージ
  const displayUser = apiUser || authUser;
  const displayName = apiUser?.username || authUser?.name;
  const displayEmail = apiUser?.email || authUser?.email;
  const emailVerified = authUser?.emailVerified || false;

  const updateMutation = useMutation({
    mutationFn: (data: UpdateMeRequest) => userService.updateMe(data),
    // UpdateMeResponse は GetMeResponse と同じ形なので、そのままキャッシュに反映する
    onSuccess: (response) => {
      queryClient.setQueryData<GetMeResponse>(ME_QUERY_KEY, response);
      setIsEditing(false);
      editForm.resetFields();
      toast.success('Profile updated successfully');
    },
    onError: (error: Error) => toast.error(error.message),
  });

  const resendVerificationMutation = useMutation({
    mutationFn: () => userService.sendMeVerifyEmail(),
    onSuccess: () => toast.success('Verification email sent'),
    onError: (error: Error) => toast.error(error.message),
  });

  // 編集開始時の処理
  const handleEditStart = () => {
    setIsEditing(true);
    editForm.setFieldsValue({
      name: displayName || '',
      email: displayEmail || '',
    });
  };

  // 編集キャンセル
  const handleEditCancel = () => {
    setIsEditing(false);
    editForm.resetFields();
  };

  // 編集保存
  const handleEditSubmit = (values: ProfileFormValues) => {
    updateMutation.mutate({ username: values.name, email: values.email });
  };

  // メール再送信（確認メール）
  const handleResendVerification = () => {
    resendVerificationMutation.mutate();
  };

  return (
    <div className="space-y-6">
      {/* ヘッダーセクション */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">My Profile</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">View and manage your account information</p>
        </div>
        <div className="flex items-center gap-3">
          {!isEditing && (
            <Button
              type="primary"
              icon={<EditOutlined />}
              onClick={handleEditStart}
              style={{ borderRadius: '8px' }}
            >
              Edit My
            </Button>
          )}
        </div>
      </div>

      {/* プロファイル情報カード */}
      {isLoading ? (
        <div className="flex justify-center py-12">
          <Spin size="large" tip="Loading profile..." />
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* メイン情報カード */}
        <Card className="lg:col-span-2 shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <div className="flex items-start gap-4 mb-6">
            <div className="h-20 w-20 rounded-full bg-blue-100 flex items-center justify-center">
              {displayName ? (
                <span className="text-blue-600 font-bold text-2xl">
                  {displayName.charAt(0).toUpperCase()}
                </span>
              ) : (
                <UserOutlined style={{ fontSize: '32px', color: '#3b82f6' }} />
              )}
            </div>
            <div>
              <h3 className="text-xl font-semibold text-gray-900 dark:text-white">{displayName || 'User'}</h3>
              <p className="text-gray-500 dark:text-gray-400">{displayEmail || 'user@example.com'}</p>
              <div className="mt-2">
                {emailVerified ? (
                  <Tag color="green" icon={<SafetyOutlined />} className="mt-1">
                    Email Verified
                  </Tag>
                ) : (
                  <Tag color="orange" icon={<SafetyOutlined />} className="mt-1">
                    Email Not Verified
                  </Tag>
                )}
              </div>
            </div>
          </div>

          <Divider />

          <Descriptions column={1} bordered size="small" className="profile-descriptions">
            <Descriptions.Item label="User ID">
              <code className="text-sm bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">
                {displayUser?.id || 'N/A'}
              </code>
            </Descriptions.Item>
            <Descriptions.Item label="Email">
              <div className="flex items-center gap-2">
                <MailOutlined />
                <span>{displayEmail || 'N/A'}</span>
              </div>
            </Descriptions.Item>
            <Descriptions.Item label="Email Verification">
              <div className="flex items-center gap-2">
                {emailVerified ? (
                  <>
                    <span className="text-green-600 dark:text-green-400">✓ Verified</span>
                  </>
                ) : (
                  <>
                    <span className="text-orange-600 dark:text-orange-400">✗ Not Verified</span>
                    <Button
                      type="link"
                      size="small"
                      onClick={handleResendVerification}
                      loading={resendVerificationMutation.isPending}
                      className="ml-2"
                    >
                      Resend Verification
                    </Button>
                  </>
                )}
              </div>
            </Descriptions.Item>
            <Descriptions.Item label="Groups">
              <div className="flex flex-wrap gap-1">
                {apiGroups && apiGroups.length > 0 ? (
                  apiGroups.map((group, index) => (
                    <Tag key={index} color="purple" className="dark:bg-purple-500/10 dark:text-purple-400">
                      {group.name || group.id}
                    </Tag>
                  ))
                ) : (
                  <span className="text-gray-400">No groups assigned</span>
                )}
              </div>
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* サイドカード - アカウント情報 */}
        <Card className="shadow-sm dark:bg-gray-800 dark:border-gray-700">
          <h3 className="text-lg font-medium mb-4 text-gray-900 dark:text-white">Account Information</h3>
          <div className="space-y-4">
            <div>
              <h4 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Account Status</h4>
              <div className="flex items-center gap-2">
                <div className={`h-2 w-2 rounded-full ${
                  apiUser?.status === 'STATUS_ACTIVE' ? 'bg-green-500' :
                  apiUser?.status === 'STATUS_INACTIVE' ? 'bg-red-500' :
                  'bg-gray-500'
                }`}></div>
                <span className={`text-sm font-medium ${
                  apiUser?.status === 'STATUS_ACTIVE' ? 'text-green-600 dark:text-green-400' :
                  apiUser?.status === 'STATUS_INACTIVE' ? 'text-red-600 dark:text-red-400' :
                  'text-gray-600 dark:text-gray-400'
                }`}>
                  {apiUser?.status ? apiUser.status.replace('STATUS_', '') : 'Unknown'}
                </span>
              </div>
            </div>
            <div>
              <h4 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Last Updated</h4>
              <p className="text-sm text-gray-700 dark:text-gray-300">
                {apiUser?.updatedAt ? new Date(apiUser.updatedAt).toLocaleDateString() : 'N/A'}
              </p>
            </div>
            <div>
              <h4 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Member Since</h4>
              <p className="text-sm text-gray-700 dark:text-gray-300">
                {apiUser?.createdAt ? new Date(apiUser.createdAt).toLocaleDateString() : 'N/A'}
              </p>
            </div>
          </div>
        </Card>
      </div>
      )}

      {/* 編集モーダル */}
      <Modal
        title="Edit My"
        open={isEditing}
        onCancel={handleEditCancel}
        maskClosable={!updateMutation.isPending}
        footer={null}
        width={500}
      >
        <Form
          form={editForm}
          layout="vertical"
          onFinish={handleEditSubmit}
        >
          <Form.Item
            name="name"
            label="Display Name"
            rules={[{ required: true, message: 'Please input your name!' }]}
          >
            <Input placeholder="Enter your name" />
          </Form.Item>
          <Form.Item
            name="email"
            label="Email Address"
            rules={[
              { required: true, message: 'Please input your email!' },
              { type: 'email', message: 'Please input a valid email!' }
            ]}
          >
            <Input type="email" placeholder="Enter your email" />
          </Form.Item>
          <Form.Item className="mb-0">
            <div className="flex justify-end gap-3">
              <Button onClick={handleEditCancel} disabled={updateMutation.isPending}>
                Cancel
              </Button>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SaveOutlined />}
                loading={updateMutation.isPending}
              >
                Save Changes
              </Button>
            </div>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
