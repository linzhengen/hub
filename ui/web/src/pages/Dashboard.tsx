import { Card, Typography } from 'antd';
import { Users, Shield, Key, FolderKanban } from 'lucide-react';

const { Title: TypographyTitle, Text } = Typography;

export function Dashboard() {
  return (
    <div className="space-y-6">
      <div>
        <TypographyTitle level={1}>Dashboard</TypographyTitle>
        <Text type="secondary">AI Hub system.</Text>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <div className="flex flex-row items-center justify-between pb-2">
            <TypographyTitle level={5} className="text-sm font-medium">Total Users</TypographyTitle>
            <Users className="h-4 w-4 text-gray-400" />
          </div>
          <div className="text-2xl font-bold">1,234</div>
        </Card>
        <Card>
          <div className="flex flex-row items-center justify-between pb-2">
            <TypographyTitle level={5} className="text-sm font-medium">Total Groups</TypographyTitle>
            <FolderKanban className="h-4 w-4 text-gray-400" />
          </div>
          <div className="text-2xl font-bold">45</div>
        </Card>
        <Card>
          <div className="flex flex-row items-center justify-between pb-2">
            <TypographyTitle level={5} className="text-sm font-medium">Total Roles</TypographyTitle>
            <Shield className="h-4 w-4 text-gray-400" />
          </div>
          <div className="text-2xl font-bold">12</div>
        </Card>
        <Card>
          <div className="flex flex-row items-center justify-between pb-2">
            <TypographyTitle level={5} className="text-sm font-medium">Total Permissions</TypographyTitle>
            <Key className="h-4 w-4 text-gray-400" />
          </div>
          <div className="text-2xl font-bold">142</div>
        </Card>
      </div>
    </div>
  );
}
