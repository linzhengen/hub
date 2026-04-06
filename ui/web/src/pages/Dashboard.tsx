import { Card, Typography, Row, Col, Progress, Statistic } from 'antd';
import { Users, Shield, Key, FolderKanban, TrendingUp, TrendingDown, Activity, BarChart3 } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';

const { Title: TypographyTitle, Text } = Typography;

// サンプルデータ
const userGrowthData = [
  { month: 'Jan', users: 400 },
  { month: 'Feb', users: 600 },
  { month: 'Mar', users: 800 },
  { month: 'Apr', users: 1000 },
  { month: 'May', users: 1200 },
  { month: 'Jun', users: 1400 },
  { month: 'Jul', users: 1600 },
];

const permissionDistributionData = [
  { name: 'Read', value: 65 },
  { name: 'Write', value: 25 },
  { name: 'Delete', value: 10 },
];

const recentActivities = [
  { id: 1, user: 'John Doe', action: 'Created new role', time: '2 hours ago' },
  { id: 2, user: 'Jane Smith', action: 'Updated permissions', time: '4 hours ago' },
  { id: 3, user: 'Bob Johnson', action: 'Added new user', time: '6 hours ago' },
  { id: 4, user: 'Alice Brown', action: 'Modified group settings', time: '1 day ago' },
];

export function Dashboard() {
  return (
    <div className="space-y-6">
      {/* ヘッダー */}
      <div className="flex justify-between items-center">
        <div>
          <TypographyTitle level={1} className="!mb-2 !text-gray-900 dark:!text-white">Dashboard</TypographyTitle>
          <Text className="text-gray-500 dark:text-gray-400">Welcome back! Here's what's happening with your AI Hub system.</Text>
        </div>
        <div className="flex gap-3">
          <button className="px-4 py-2 bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 rounded-lg font-medium hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors">
            Generate Report
          </button>
          <button className="px-4 py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors">
            Add User
          </button>
        </div>
      </div>

      {/* 統計カード */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card className="shadow-sm hover:shadow-md transition-shadow dark:bg-gray-800 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <Text className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Users</Text>
                <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">1,234</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className="h-4 w-4 text-green-500 dark:text-green-400" />
                  <span className="text-sm text-green-600 dark:text-green-400">+12.5%</span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-blue-50 dark:bg-blue-900/20">
                <Users className="h-6 w-6 text-blue-600 dark:text-blue-400" />
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card className="shadow-sm hover:shadow-md transition-shadow dark:bg-gray-800 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <Text className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Groups</Text>
                <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">45</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className="h-4 w-4 text-green-500 dark:text-green-400" />
                  <span className="text-sm text-green-600 dark:text-green-400">+5.2%</span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-green-50 dark:bg-green-900/20">
                <FolderKanban className="h-6 w-6 text-green-600 dark:text-green-400" />
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card className="shadow-sm hover:shadow-md transition-shadow dark:bg-gray-800 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <Text className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Roles</Text>
                <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">12</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingDown className="h-4 w-4 text-red-500 dark:text-red-400" />
                  <span className="text-sm text-red-600 dark:text-red-400">-2.1%</span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-purple-50 dark:bg-purple-900/20">
                <Shield className="h-6 w-6 text-purple-600 dark:text-purple-400" />
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card className="shadow-sm hover:shadow-md transition-shadow dark:bg-gray-800 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <Text className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Permissions</Text>
                <div className="text-2xl font-bold mt-1 text-gray-900 dark:text-white">142</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className="h-4 w-4 text-green-500 dark:text-green-400" />
                  <span className="text-sm text-green-600 dark:text-green-400">+8.7%</span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-orange-50 dark:bg-orange-900/20">
                <Key className="h-6 w-6 text-orange-600 dark:text-orange-400" />
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* チャートセクション */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <div className="flex items-center gap-2">
                <Activity className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                <span className="text-gray-900 dark:text-white">User Growth</span>
              </div>
            }
            className="shadow-sm dark:bg-gray-800 dark:border-gray-700"
          >
            <div style={{ height: '300px', minHeight: '300px' }}>
              <ResponsiveContainer width="100%" height="100%" minHeight={300} debounce={200}>
                <LineChart data={userGrowthData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="month" stroke="#64748b" />
                  <YAxis stroke="#64748b" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--tooltip-bg, #ffffff)',
                      border: '1px solid var(--tooltip-border, #e2e8f0)',
                      borderRadius: '8px',
                      color: 'var(--tooltip-text, #1e293b)'
                    }}
                  />
                  <Line
                    type="monotone"
                    dataKey="users"
                    stroke="#3b82f6"
                    strokeWidth={2}
                    dot={{ fill: '#3b82f6', strokeWidth: 2 }}
                    activeDot={{ r: 6, fill: '#3b82f6' }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card
            title={
              <div className="flex items-center gap-2">
                <BarChart3 className="h-5 w-5 text-purple-600 dark:text-purple-400" />
                <span className="text-gray-900 dark:text-white">Permission Distribution</span>
              </div>
            }
            className="shadow-sm dark:bg-gray-800 dark:border-gray-700"
          >
            <div style={{ height: '300px', minHeight: '300px' }}>
              <ResponsiveContainer width="100%" height="100%" minHeight={300} debounce={200}>
                <BarChart data={permissionDistributionData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="name" stroke="#64748b" />
                  <YAxis stroke="#64748b" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: 'var(--tooltip-bg, #ffffff)',
                      border: '1px solid var(--tooltip-border, #e2e8f0)',
                      borderRadius: '8px',
                      color: 'var(--tooltip-text, #1e293b)'
                    }}
                  />
                  <Bar
                    dataKey="value"
                    fill="#8b5cf6"
                    radius={[4, 4, 0, 0]}
                  />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </Card>
        </Col>
      </Row>

      {/* アクティビティとシステムステータス */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card
            title={
              <div className="flex items-center gap-2">
                <Activity className="h-5 w-5 text-green-600 dark:text-green-400" />
                <span className="text-gray-900 dark:text-white">Recent Activities</span>
              </div>
            }
            className="shadow-sm dark:bg-gray-800 dark:border-gray-700"
          >
            <div className="space-y-4">
              {recentActivities.map((activity) => (
                <div key={activity.id} className="flex items-center justify-between p-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors">
                  <div>
                    <div className="font-medium text-gray-900 dark:text-white">{activity.user}</div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">{activity.action}</div>
                  </div>
                  <div className="text-sm text-gray-400 dark:text-gray-500">{activity.time}</div>
                </div>
              ))}
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title={
              <div className="flex items-center gap-2">
                <Shield className="h-5 w-5 text-orange-600 dark:text-orange-400" />
                <span className="text-gray-900 dark:text-white">System Status</span>
              </div>
            }
            className="shadow-sm dark:bg-gray-800 dark:border-gray-700"
          >
            <div className="space-y-6">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium text-gray-900 dark:text-white">API Response Time</span>
                  <span className="text-sm font-medium text-blue-600 dark:text-blue-400">85%</span>
                </div>
                <Progress percent={85} strokeColor="#3b82f6" />
              </div>

              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium text-gray-900 dark:text-white">Database Health</span>
                  <span className="text-sm font-medium text-green-600 dark:text-green-400">98%</span>
                </div>
                <Progress percent={98} strokeColor="#10b981" />
              </div>

              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium text-gray-900 dark:text-white">Memory Usage</span>
                  <span className="text-sm font-medium text-orange-600 dark:text-orange-400">72%</span>
                </div>
                <Progress percent={72} strokeColor="#f59e0b" />
              </div>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
}
