import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { roleService, Role, CreateRoleRequest, UpdateRoleRequest } from '@/services/role';
import { permissionService, Permission } from '@/services/permission';
import { Button, buttonVariants } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Plus, Edit, Trash2, Key } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';

export function Roles() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [managingPermissionsRole, setManagingPermissionsRole] = useState<Role | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['roles'],
    queryFn: () => roleService.listRoles(),
  });

  const { data: permissionsData } = useQuery({
    queryKey: ['permissions'],
    queryFn: () => permissionService.listPermissions(),
  });

  const createMutation = useMutation({
    mutationFn: roleService.createRole,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setIsCreateOpen(false);
      toast.success('Role created successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateRoleRequest }) => roleService.updateRole(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setEditingRole(null);
      toast.success('Role updated successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const deleteMutation = useMutation({
    mutationFn: roleService.deleteRole,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      toast.success('Role deleted successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const assignPermissionMutation = useMutation({
    mutationFn: ({ id, permissionId }: { id: string; permissionId: string }) => roleService.assignPermission(id, { permissionId }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      // Update managingPermissionsRole state immediately
      if (managingPermissionsRole && managingPermissionsRole.id === variables.id) {
        setManagingPermissionsRole({
          ...managingPermissionsRole,
          permissionIds: [...managingPermissionsRole.permissionIds, variables.permissionId]
        });
      }
      toast.success('Permission assigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const unassignPermissionMutation = useMutation({
    mutationFn: ({ id, permissionId, currentPermissionIds }: { id: string; permissionId: string; currentPermissionIds: string[] }) =>
      roleService.removePermissionsFromRole(id, { permissionIds: currentPermissionIds.filter(pId => pId !== permissionId) }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      // Update managingPermissionsRole state immediately
      if (managingPermissionsRole && managingPermissionsRole.id === variables.id) {
        setManagingPermissionsRole({
          ...managingPermissionsRole,
          permissionIds: managingPermissionsRole.permissionIds.filter(id => id !== variables.permissionId)
        });
      }
      toast.success('Permission unassigned successfully');
    },
    onError: (error: any) => toast.error(error.message),
  });

  const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    createMutation.mutate({
      name: formData.get('name') as string,
      description: formData.get('description') as string,
    });
  };

  const handleUpdate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editingRole) return;
    const formData = new FormData(e.currentTarget);
    updateMutation.mutate({
      id: editingRole.id,
      data: {
        name: formData.get('name') as string,
        description: formData.get('description') as string,
      }
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Roles</h2>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger className={cn(buttonVariants())}>
            <Plus className="mr-2 h-4 w-4" /> Add Role
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Role</DialogTitle>
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
              <TableHead>Permissions</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center">Loading...</TableCell>
              </TableRow>
            ) : error ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-red-500">Failed to load roles</TableCell>
              </TableRow>
            ) : data?.roles?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center">No roles found</TableCell>
              </TableRow>
            ) : (
              data?.roles?.map((role) => (
                <TableRow key={role.id}>
                  <TableCell className="font-medium">{role.name}</TableCell>
                  <TableCell>{role.description}</TableCell>
                  <TableCell>{role.permissionIds?.length || 0} permission(s)</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button variant="ghost" size="icon" onClick={() => setManagingPermissionsRole(role)}>
                        <Key className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => setEditingRole(role)}>
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => {
                        if (confirm('Are you sure you want to delete this role?')) {
                          deleteMutation.mutate(role.id);
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

      <Dialog open={!!editingRole} onOpenChange={(open) => !open && setEditingRole(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Role</DialogTitle>
          </DialogHeader>
          {editingRole && (
            <form onSubmit={handleUpdate} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-name">Name</Label>
                <Input id="edit-name" name="name" defaultValue={editingRole.name} required />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-description">Description</Label>
                <Input id="edit-description" name="description" defaultValue={editingRole.description} />
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

      <Dialog open={!!managingPermissionsRole} onOpenChange={(open) => !open && setManagingPermissionsRole(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Manage Permissions for {managingPermissionsRole?.name}</DialogTitle>
          </DialogHeader>
          {managingPermissionsRole && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium">Current Permissions</h3>
                {managingPermissionsRole.permissionIds.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No permissions assigned</p>
                ) : (
                  <div className="mt-2 space-y-2">
                    {managingPermissionsRole.permissionIds.map((permissionId) => {
                      const permission = permissionsData?.permissions?.find(p => p.id === permissionId);
                      return permission ? (
                        <div key={permissionId} className="flex items-center justify-between rounded-md border px-3 py-2">
                          <span>{permission.name} ({permission.verb} on {permission.resourceId})</span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              if (confirm(`Are you sure you want to remove ${permission.name} from this role?`)) {
                                unassignPermissionMutation.mutate({
                                  id: managingPermissionsRole.id,
                                  permissionId: permission.id,
                                  currentPermissionIds: managingPermissionsRole.permissionIds
                                });
                              }
                            }}
                            disabled={unassignPermissionMutation.isPending && unassignPermissionMutation.variables?.id === managingPermissionsRole.id && unassignPermissionMutation.variables?.permissionId === permission.id}
                          >
                            {unassignPermissionMutation.isPending && unassignPermissionMutation.variables?.id === managingPermissionsRole.id && unassignPermissionMutation.variables?.permissionId === permission.id ? 'Removing...' : 'Remove'}
                          </Button>
                        </div>
                      ) : null;
                    })}
                  </div>
                )}
              </div>

              <div>
                <h3 className="text-lg font-medium">Available Permissions</h3>
                {permissionsData?.permissions?.filter(permission => !managingPermissionsRole.permissionIds.includes(permission.id)).length === 0 ? (
                  <p className="text-sm text-muted-foreground">No available permissions</p>
                ) : (
                  <div className="mt-2 space-y-2">
                    {permissionsData?.permissions
                      ?.filter(permission => !managingPermissionsRole.permissionIds.includes(permission.id))
                      .map((permission) => (
                        <div key={permission.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                          <span>{permission.name} ({permission.verb} on {permission.resourceId})</span>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              assignPermissionMutation.mutate({
                                id: managingPermissionsRole.id,
                                permissionId: permission.id
                              });
                            }}
                            disabled={assignPermissionMutation.isPending && assignPermissionMutation.variables?.id === managingPermissionsRole.id && assignPermissionMutation.variables?.permissionId === permission.id}
                          >
                            {assignPermissionMutation.isPending && assignPermissionMutation.variables?.id === managingPermissionsRole.id && assignPermissionMutation.variables?.permissionId === permission.id ? 'Assigning...' : 'Assign'}
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
