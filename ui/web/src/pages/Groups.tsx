import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { groupService, Group, CreateGroupRequest, UpdateGroupRequest, GroupStatus } from '@/services/group';
import { roleService, Role } from '@/services/role';
import { userService, User } from '@/services/user';
import { Button, buttonVariants } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Plus, Edit, Trash2, Key, Users as UsersIcon } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';

export function Groups() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [managingRolesGroup, setManagingRolesGroup] = useState<Group | null>(null);
  const [managingUsersGroup, setManagingUsersGroup] = useState<Group | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['groups'],
    queryFn: () => groupService.listGroups(),
  });

  const { data: rolesData } = useQuery({
    queryKey: ['roles'],
    queryFn: () => roleService.listRoles(),
  });

  const { data: usersData } = useQuery({
    queryKey: ['users'],
    queryFn: () => userService.listUsers(),
  });

  const createMutation = useMutation({
    mutationFn: groupService.createGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setIsCreateOpen(false);
      toast.success('Group created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateGroupRequest }) => groupService.updateGroup(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setEditingGroup(null);
      toast.success('Group updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: groupService.deleteGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      toast.success('Group deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const assignRoleMutation = useMutation({
    mutationFn: ({ id, roleId }: { id: string; roleId: string }) => groupService.assignRole(id, { roleId }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingRolesGroup state immediately
      if (managingRolesGroup && managingRolesGroup.id === variables.id) {
        setManagingRolesGroup({
          ...managingRolesGroup,
          roleIds: [...managingRolesGroup.roleIds, variables.roleId]
        });
      }
      toast.success('Role assigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const unassignRoleMutation = useMutation({
    mutationFn: ({ id, roleId, currentRoleIds }: { id: string; roleId: string; currentRoleIds: string[] }) =>
      groupService.assignRolesToGroup(id, { roleIds: currentRoleIds.filter(rId => rId !== roleId) }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingRolesGroup state immediately
      if (managingRolesGroup && managingRolesGroup.id === variables.id) {
        setManagingRolesGroup({
          ...managingRolesGroup,
          roleIds: managingRolesGroup.roleIds.filter(id => id !== variables.roleId)
        });
      }
      toast.success('Role unassigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const addUsersToGroupMutation = useMutation({
    mutationFn: ({ id, userIds }: { id: string; userIds: string[] }) => groupService.addUsersToGroup(id, { userIds }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingUsersGroup state immediately
      if (managingUsersGroup && managingUsersGroup.id === variables.id) {
        // Note: group doesn't have userIds field in response, we'll just invalidate
      }
      toast.success('Users added successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const removeUsersFromGroupMutation = useMutation({
    mutationFn: ({ id, userIds }: { id: string; userIds: string[] }) => groupService.removeUsersFromGroup(id, { userIds }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      // Update managingUsersGroup state immediately
      if (managingUsersGroup && managingUsersGroup.id === variables.id) {
        // Note: group doesn't have userIds field in response, we'll just invalidate
      }
      toast.success('Users removed successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    createMutation.mutate({
      name: formData.get('name') as string,
      description: formData.get('description') as string,
      status: (formData.get('status') as GroupStatus) || 'STATUS_UNSPECIFIED',
    });
  };

  const handleUpdate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editingGroup) return;
    const formData = new FormData(e.currentTarget);
    const data: UpdateGroupRequest = {
      name: formData.get('name') as string,
      description: formData.get('description') as string,
    };
    const status = formData.get('status') as GroupStatus;
    if (status) {
      data.status = status;
    }
    updateMutation.mutate({
      id: editingGroup.id,
      data
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Groups</h2>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger className={cn(buttonVariants())}>
            <Plus className="mr-2 h-4 w-4" /> Add Group
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Group</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input id="name" name="name" required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input id="description" name="description" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="status">Status</Label>
                <select
                  id="status"
                  name="status"
                  defaultValue="STATUS_UNSPECIFIED"
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="STATUS_UNSPECIFIED">Unspecified</option>
                  <option value="STATUS_ACTIVE">Active</option>
                  <option value="STATUS_INACTIVE">Inactive</option>
                </select>
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
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Roles</TableHead>
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
                <TableCell colSpan={5} className="text-center text-red-500">Failed to load groups</TableCell>
              </TableRow>
            ) : data?.groups?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center">No groups found</TableCell>
              </TableRow>
            ) : (
              data?.groups?.map((group) => (
                <TableRow key={group.id}>
                  <TableCell className="font-medium">{group.name}</TableCell>
                  <TableCell>{group.description}</TableCell>
                  <TableCell>
                    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ${group.status === 'STATUS_ACTIVE' ? 'bg-green-100 text-green-800' : group.status === 'STATUS_INACTIVE' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800'}`}>
                      {group.status === 'STATUS_ACTIVE' ? 'Active' : group.status === 'STATUS_INACTIVE' ? 'Inactive' : 'Unspecified'}
                    </span>
                  </TableCell>
                  <TableCell>
                    {group.roleIds?.length || 0} role(s)
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="icon" onClick={() => setManagingRolesGroup(group)}>
                        <Key className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => setManagingUsersGroup(group)}>
                        <UsersIcon className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => setEditingGroup(group)}>
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm('Are you sure you want to delete this group?')) {
                          deleteMutation.mutate(group.id);
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

      <Dialog open={!!editingGroup} onOpenChange={(open) => !open && setEditingGroup(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Group</DialogTitle>
          </DialogHeader>
          {editingGroup && (
            <form onSubmit={handleUpdate} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-name">Name</Label>
                <Input id="edit-name" name="name" defaultValue={editingGroup.name} required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-description">Description</Label>
                <Input id="edit-description" name="description" defaultValue={editingGroup.description} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-status">Status</Label>
                <select
                  id="edit-status"
                  name="status"
                  defaultValue={editingGroup.status}
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="STATUS_UNSPECIFIED">Unspecified</option>
                  <option value="STATUS_ACTIVE">Active</option>
                  <option value="STATUS_INACTIVE">Inactive</option>
                </select>
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

      <Dialog open={!!managingRolesGroup} onOpenChange={(open) => !open && setManagingRolesGroup(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Manage Roles for {managingRolesGroup?.name}</DialogTitle>
          </DialogHeader>
          {managingRolesGroup && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium">Current Roles</h3>
                {managingRolesGroup.roleIds.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No roles assigned</p>
                ) : (
                  <div className="mt-2 space-y-2">
                    {managingRolesGroup.roleIds.map((roleId) => {
                      const role = rolesData?.roles?.find(r => r.id === roleId);
                      return role ? (
                        <div key={roleId} className="flex items-center justify-between rounded-md border px-3 py-2">
                          <span>{role.name}</span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              if (confirm(`Are you sure you want to remove ${role.name} from this group?`)) {
                                unassignRoleMutation.mutate({
                                  id: managingRolesGroup.id,
                                  roleId: role.id,
                                  currentRoleIds: managingRolesGroup.roleIds
                                });
                              }
                            }}
                            disabled={unassignRoleMutation.isPending && unassignRoleMutation.variables?.id === managingRolesGroup.id && unassignRoleMutation.variables?.roleId === role.id}
                          >
                            {unassignRoleMutation.isPending && unassignRoleMutation.variables?.id === managingRolesGroup.id && unassignRoleMutation.variables?.roleId === role.id ? 'Removing...' : 'Remove'}
                          </Button>
                        </div>
                      ) : null;
                    })}
                  </div>
                )}
              </div>

              <div>
                <h3 className="text-lg font-medium">Available Roles</h3>
                {rolesData?.roles?.filter(role => !managingRolesGroup.roleIds.includes(role.id)).length === 0 ? (
                  <p className="text-sm text-muted-foreground">No available roles</p>
                ) : (
                  <div className="mt-2 space-y-2">
                    {rolesData?.roles
                      ?.filter(role => !managingRolesGroup.roleIds.includes(role.id))
                      .map((role) => (
                        <div key={role.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                          <span>{role.name}</span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              assignRoleMutation.mutate({
                                id: managingRolesGroup.id,
                                roleId: role.id
                              });
                            }}
                            disabled={assignRoleMutation.isPending && assignRoleMutation.variables?.id === managingRolesGroup.id && assignRoleMutation.variables?.roleId === role.id}
                          >
                            {assignRoleMutation.isPending && assignRoleMutation.variables?.id === managingRolesGroup.id && assignRoleMutation.variables?.roleId === role.id ? 'Assigning...' : 'Assign'}
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

      <Dialog open={!!managingUsersGroup} onOpenChange={(open) => !open && setManagingUsersGroup(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Manage Users for {managingUsersGroup?.name}</DialogTitle>
          </DialogHeader>
          {managingUsersGroup && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium">Current Users</h3>
                <p className="text-sm text-muted-foreground">User list not available in group response</p>
              </div>

              <div>
                <h3 className="text-lg font-medium">Add Users</h3>
                <div className="mt-2 space-y-2">
                  {usersData?.users?.map((user) => (
                    <div key={user.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                      <span>{user.username} ({user.email})</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          addUsersToGroupMutation.mutate({
                            id: managingUsersGroup.id,
                            userIds: [user.id]
                          });
                        }}
                        disabled={addUsersToGroupMutation.isPending && addUsersToGroupMutation.variables?.id === managingUsersGroup.id && addUsersToGroupMutation.variables?.userIds?.includes(user.id)}
                      >
                        {addUsersToGroupMutation.isPending && addUsersToGroupMutation.variables?.id === managingUsersGroup.id && addUsersToGroupMutation.variables?.userIds?.includes(user.id) ? 'Adding...' : 'Add'}
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
