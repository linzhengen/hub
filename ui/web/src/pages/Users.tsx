import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userService, User } from '@/services/user';
import { groupService, Group, ListGroupsResponse } from '@/services/group';
import { Button, buttonVariants } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Plus, Edit, Trash2, Users as UsersIcon } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

import { toast } from 'sonner';
import { cn } from '@/lib/utils';

export function Users() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [managingGroupsUser, setManagingGroupsUser] = useState<User | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['users'],
    queryFn: () => userService.listUsers(),
  });

  const { data: groupsData } = useQuery<ListGroupsResponse>({
    queryKey: ['groups'],
    queryFn: () => groupService.listGroups(),
  });

  // Helper function to get group names from group IDs
  const getGroupNames = (groupIds: string[]): string[] => {
    if (!groupsData?.groups) return [];
    return groupIds
      .map(id => groupsData.groups.find(g => g.id === id)?.name)
      .filter((name): name is string => name !== undefined);
  };

  const createMutation = useMutation({
    mutationFn: userService.createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setIsCreateOpen(false);
      toast.success('User created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<User> }) => userService.updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setEditingUser(null);
      toast.success('User updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: userService.deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      toast.success('User deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const assignGroupMutation = useMutation({
    mutationFn: ({ userId, groupId }: { userId: string; groupId: string }) =>
      userService.assignGroup(userId, groupId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      // Update the managingGroupsUser state to reflect the change immediately
      if (managingGroupsUser && managingGroupsUser.id === variables.userId) {
        setManagingGroupsUser({
          ...managingGroupsUser,
          groupIds: [...managingGroupsUser.groupIds, variables.groupId]
        });
      }
      toast.success('Group assigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const unassignGroupMutation = useMutation({
    mutationFn: ({ userId, groupId }: { userId: string; groupId: string }) =>
      userService.unassignGroup(userId, groupId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      // Update the managingGroupsUser state to reflect the change immediately
      if (managingGroupsUser && managingGroupsUser.id === variables.userId) {
        setManagingGroupsUser({
          ...managingGroupsUser,
          groupIds: managingGroupsUser.groupIds.filter(id => id !== variables.groupId)
        });
      }
      toast.success('Group unassigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const groupIds = formData.getAll('groupIds') as string[];
    createMutation.mutate({
      username: formData.get('username') as string,
      email: formData.get('email') as string,
      password: formData.get('password') as string,
      groupIds: groupIds,
    });
  };

  const handleUpdate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editingUser) return;
    const formData = new FormData(e.currentTarget);
    const data: any = {
      username: formData.get('username') as string,
      email: formData.get('email') as string,
    };
    const password = formData.get('password') as string;
    if (password) data.password = password;
    const status = formData.get('status') as string;
    if (status) data.status = status;
    const groupIds = formData.getAll('groupIds') as string[];
    if (groupIds.length > 0) {
      data.groupIds = groupIds;
    }
    updateMutation.mutate({
      id: editingUser.id,
      data
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Users</h2>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger className={cn(buttonVariants())}>
            <Plus className="mr-2 h-4 w-4" /> Add User
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New User</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <Input id="username" name="username" required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input id="email" name="email" type="email" required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input id="password" name="password" type="password" required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="groupIds">Groups (optional)</Label>
                <select
                  id="groupIds"
                  name="groupIds"
                  multiple
                  className="flex h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {groupsData?.groups?.map((group) => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-muted-foreground">Hold Ctrl (Cmd on Mac) to select multiple groups</p>
              </div>
              <div className="flex justify-end">
                <Button type="submit" disabled={createMutation.isPending}>
                  {createMutation.isPending ? 'Creating...' : 'Create'}
                </Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Groups</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center">Loading...</TableCell>
              </TableRow>
            ) : error ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-red-500">Failed to load users</TableCell>
              </TableRow>
            ) : data?.users?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center">No users found</TableCell>
              </TableRow>
            ) : (
              data?.users?.map((user) => (
                <TableRow key={user.id}>
                  <TableCell className="font-medium">{user.username}</TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>
                    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${user.status === 'STATUS_ACTIVE' ? 'bg-green-100 text-green-800' : user.status === 'STATUS_INACTIVE' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800'}`}>
                      {user.status === 'STATUS_ACTIVE' ? 'Active' : user.status === 'STATUS_INACTIVE' ? 'Inactive' : 'Unspecified'}
                    </span>
                  </TableCell>
                  <TableCell>
                    {getGroupNames(user.groupIds).join(', ') || 'None'}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="icon" onClick={() => setManagingGroupsUser(user)}>
                        <UsersIcon className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => setEditingUser(user)}>
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm('Are you sure you want to delete this user?')) {
                          deleteMutation.mutate(user.id);
                        }
                      }}>
                        <Trash2 className="h-4 w-4 text-red-500" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <Dialog open={!!editingUser} onOpenChange={(open) => !open && setEditingUser(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
          </DialogHeader>
          {editingUser && (
            <form onSubmit={handleUpdate} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-username">Username</Label>
                <Input id="edit-username" name="username" defaultValue={editingUser.username} required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-email">Email</Label>
                <Input id="edit-email" name="email" type="email" defaultValue={editingUser.email} required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-password">Password (leave empty to keep unchanged)</Label>
                <Input id="edit-password" name="password" type="password" placeholder="••••••••" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-status">Status</Label>
                <select
                  id="edit-status"
                  name="status"
                  defaultValue={editingUser.status}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="STATUS_UNSPECIFIED">Unspecified</option>
                  <option value="STATUS_ACTIVE">Active</option>
                  <option value="STATUS_INACTIVE">Inactive</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-groupIds">Groups (optional)</Label>
                <select
                  id="edit-groupIds"
                  name="groupIds"
                  multiple
                  defaultValue={editingUser.groupIds}
                  className="flex h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {groupsData?.groups?.map((group) => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-muted-foreground">Hold Ctrl (Cmd on Mac) to select multiple groups</p>
              </div>
              <div className="flex justify-end">
                <Button type="submit" disabled={updateMutation.isPending}>
                  {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
                </Button>
              </div>
            </form>
          )}
        </DialogContent>
      </Dialog>

      <Dialog open={!!managingGroupsUser} onOpenChange={(open) => !open && setManagingGroupsUser(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Manage Groups for {managingGroupsUser?.username}</DialogTitle>
          </DialogHeader>
          {managingGroupsUser && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium">Current Groups</h3>
                {managingGroupsUser.groupIds.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No groups assigned</p>
                ) : (
                  <div className="mt-2 space-y-2">
                    {managingGroupsUser.groupIds.map((groupId) => {
                      const group = groupsData?.groups?.find(g => g.id === groupId);
                      return group ? (
                        <div key={groupId} className="flex items-center justify-between rounded-md border px-3 py-2">
                          <span>{group.name}</span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              if (confirm(`Are you sure you want to remove ${group.name} from this user?`)) {
                                unassignGroupMutation.mutate({
                                  userId: managingGroupsUser.id,
                                  groupId: group.id
                                });
                              }
                            }}
                            disabled={unassignGroupMutation.isPending && unassignGroupMutation.variables?.groupId === group.id}
                          >
                            {unassignGroupMutation.isPending && unassignGroupMutation.variables?.groupId === group.id ? 'Removing...' : 'Remove'}
                          </Button>
                        </div>
                      ) : null;
                    })}
                  </div>
                )}
              </div>

              <div>
                <h3 className="text-lg font-medium">Available Groups</h3>
                {groupsData?.groups?.filter(group => !managingGroupsUser.groupIds.includes(group.id)).length === 0 ? (
                  <p className="text-sm text-muted-foreground">No available groups</p>
                ) : (
                  <div className="mt-2 space-y-2">
                    {groupsData?.groups
                      ?.filter(group => !managingGroupsUser.groupIds.includes(group.id))
                      .map((group) => (
                        <div key={group.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                          <span>{group.name}</span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              assignGroupMutation.mutate({
                                userId: managingGroupsUser.id,
                                groupId: group.id
                              });
                            }}
                            disabled={assignGroupMutation.isPending && assignGroupMutation.variables?.groupId === group.id}
                          >
                            {assignGroupMutation.isPending && assignGroupMutation.variables?.groupId === group.id ? 'Assigning...' : 'Assign'}
                          </Button>
                        </div>
                      ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
