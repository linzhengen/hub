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
          <TypographyTitle level={1} style={{ marginBottom: '8px', color: '#1e293b' }}>Dashboard</TypographyTitle>
          <Text type="secondary" style={{ color: '#64748b' }}>Welcome back! Here's what's happening with your AI Hub system.</Text>
        </div>
        <div className="flex gap-3">
          <button className="px-4 py-2 bg-blue-50 text-blue-600 rounded-lg font-medium hover:bg-blue-100 transition-colors">
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
          <Card className="shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <Text type="secondary" className="text-sm font-medium">Total Users</Text>
                <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>1,234</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className="h-4 w-4 text-green-500" />
                  <span className="text-sm text-green-600">+12.5%</span>
                  <span className="text-sm text-gray-500">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-blue-50">
                <Users className="h-6 w-6 text-blue-600" />
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card className="shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <Text type="secondary" className="text-sm font-medium">Total Groups</Text>
                <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>45</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className="h-4 w-4 text-green-500" />
                  <span className="text-sm text-green-600">+5.2%</span>
                  <span className="text-sm text-gray-500">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-green-50">
                <FolderKanban className="h-6 w-6 text-green-600" />
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card className="shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <Text type="secondary" className="text-sm font-medium">Total Roles</Text>
                <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>12</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingDown className="h-4 w-4 text-red-500" />
                  <span className="text-sm text-red-600">-2.1%</span>
                  <span className="text-sm text-gray-500">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-purple-50">
                <Shield className="h-6 w-6 text-purple-600" />
              </div>
            </div>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card className="shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <Text type="secondary" className="text-sm font-medium">Total Permissions</Text>
                <div className="text-2xl font-bold mt-1" style={{ color: '#1e293b' }}>142</div>
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className="h-4 w-4 text-green-500" />
                  <span className="text-sm text-green-600">+8.7%</span>
                  <span className="text-sm text-gray-500">from last month</span>
                </div>
              </div>
              <div className="p-3 rounded-lg bg-orange-50">
                <Key className="h-6 w-6 text-orange-600" />
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
                <Activity className="h-5 w-5 text-blue-600" />
                <span style={{ color: '#1e293b' }}>User Growth</span>
              </div>
            }
            className="shadow-sm"
          >
            <div style={{ height: '300px', minHeight: '300px' }}>
              <ResponsiveContainer width="100%" height="100%" minHeight={300} debounce={200}>
                <LineChart data={userGrowthData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="month" stroke="#64748b" />
                  <YAxis stroke="#64748b" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#ffffff',
                      border: '1px solid #e2e8f0',
                      borderRadius: '8px'
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
                <BarChart3 className="h-5 w-5 text-purple-600" />
                <span style={{ color: '#1e293b' }}>Permission Distribution</span>
              </div>
            }
            className="shadow-sm"
          >
            <div style={{ height: '300px', minHeight: '300px' }}>
              <ResponsiveContainer width="100%" height="100%" minHeight={300} debounce={200}>
                <BarChart data={permissionDistributionData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="name" stroke="#64748b" />
                  <YAxis stroke="#64748b" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#ffffff',
                      border: '1px solid #e2e8f0',
                      borderRadius: '8px'
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
                <Activity className="h-5 w-5 text-green-600" />
                <span style={{ color: '#1e293b' }}>Recent Activities</span>
              </div>
            }
            className="shadow-sm"
          >
            <div className="space-y-4">
              {recentActivities.map((activity) => (
                <div key={activity.id} className="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg transition-colors">
                  <div>
                    <div className="font-medium" style={{ color: '#1e293b' }}>{activity.user}</div>
                    <div className="text-sm" style={{ color: '#64748b' }}>{activity.action}</div>
                  </div>
                  <div className="text-sm" style={{ color: '#94a3b8' }}>{activity.time}</div>
                </div>
              ))}
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title={
              <div className="flex items-center gap-2">
                <Shield className="h-5 w-5 text-orange-600" />
                <span style={{ color: '#1e293b' }}>System Status</span>
              </div>
            }
            className="shadow-sm"
          >
            <div className="space-y-6">
              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium" style={{ color: '#1e293b' }}>API Response Time</span>
                  <span className="text-sm font-medium" style={{ color: '#3b82f6' }}>85%</span>
                </div>
                <Progress percent={85} strokeColor="#3b82f6" />
              </div>

              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium" style={{ color: '#1e293b' }}>Database Health</span>
                  <span className="text-sm font-medium" style={{ color: '#10b981' }}>98%</span>
                </div>
                <Progress percent={98} strokeColor="#10b981" />
              </div>

              <div>
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium" style={{ color: '#1e293b' }}>Memory Usage</span>
                  <span className="text-sm font-medium" style={{ color: '#f59e0b' }}>72%</span>
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
