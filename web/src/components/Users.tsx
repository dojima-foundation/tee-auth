'use client';

import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { User as UserIcon, Building, Edit, Trash2 } from 'lucide-react';
import { useSnackbar } from '@/components/ui/snackbar';
import { gauthApi, CreateOrganizationRequest } from '@/services/gauthApi';
import { useAuth } from '@/lib/auth-context';
import CreateUserDialog from './CreateUserDialog';
import {
    fetchUsers,
    createUser,
    selectUsers,
    selectUsersLoading,
    selectUsersError,
    selectUsersPagination,
    type User
} from '@/store/usersSlice';
import { selectOrganizationId, selectAuthUser, selectAuthSession } from '@/store/authSlice';

export default function Users() {
    const dispatch = useAppDispatch();
    const { isAuthenticated } = useAuth();

    // Get users data from Redux store
    const users = useAppSelector(selectUsers);
    const loading = useAppSelector(selectUsersLoading);
    // const error = useAppSelector(selectUsersError);
    const pagination = useAppSelector(selectUsersPagination);

    // Get organization ID from auth store
    const organizationId = useAppSelector(selectOrganizationId);
    const currentUser = useAppSelector(selectAuthUser);
    const authSession = useAppSelector(selectAuthSession);

    console.log('🔍 [Users] Redux selectors:', {
        organizationId,
        currentUser,
        authSession,
        userOrganizationId: currentUser?.organization_id,
        sessionOrganizationId: (authSession as { organization_id?: string })?.organization_id
    });

    const [isCreatingOrg, setIsCreatingOrg] = useState(false);
    const [isCreatingUser, setIsCreatingUser] = useState(false);
    const { showSnackbar } = useSnackbar();

    useEffect(() => {
        console.log('🔄 [Users] useEffect triggered:', {
            organizationId,
            isAuthenticated,
            hasOrganizationId: !!organizationId,
            shouldFetch: organizationId && isAuthenticated
        });

        if (organizationId && isAuthenticated) {
            console.log('📡 [Users] Fetching users data...');
            dispatch(fetchUsers({}));
        } else {
            console.log('⏸️ [Users] Not fetching users - missing requirements:', {
                hasOrganizationId: !!organizationId,
                isAuthenticated
            });
        }
    }, [dispatch, organizationId, isAuthenticated]);

    const handleCreateOrganization = async () => {
        try {
            setIsCreatingOrg(true);

            // Create organization with initial user
            const orgData: CreateOrganizationRequest = {
                name: 'ODEYS Organization',
                initial_user_email: 'admin@odeys.com',
                initial_user_public_key: '0x1234567890abcdef1234567890abcdef12345678'
            };

            const response = await gauthApi.createOrganization(orgData);

            if (response.success) {
                showSnackbar({
                    type: 'success',
                    title: 'Organization Created',
                    message: `Organization created successfully! Organization ID: ${response.data.organization.id}, User ID: ${response.data.user_id}`
                });

                // Refresh users list after creating organization
                if (organizationId) {
                    dispatch(fetchUsers({}));
                }
            } else {
                showSnackbar({
                    type: 'error',
                    title: 'Organization Creation Failed',
                    message: 'Failed to create organization'
                });
            }
        } catch (error) {
            console.error('Error creating organization:', error);
            showSnackbar({
                type: 'error',
                title: 'Organization Creation Failed',
                message: error instanceof Error ? error.message : 'Failed to create organization'
            });
        } finally {
            setIsCreatingOrg(false);
        }
    };

    const handleCreateUser = async (userData: { name: string; email: string; role: string }) => {
        try {
            setIsCreatingUser(true);

            if (!organizationId) {
                throw new Error('No organization ID available');
            }

            // Call Redux action to create user
            await dispatch(createUser({
                username: userData.name,
                email: userData.email,
            })).unwrap();

            showSnackbar({
                type: 'success',
                title: 'User Created',
                message: `User "${userData.name}" created successfully!`
            });

            // Refresh users list after creating user
            dispatch(fetchUsers({}));
        } catch (error) {
            console.error('Error creating user:', error);
            showSnackbar({
                type: 'error',
                title: 'User Creation Failed',
                message: error instanceof Error ? error.message : 'Failed to create user'
            });
        } finally {
            setIsCreatingUser(false);
        }
    };

    const handleEditUser = async (user: User) => {
        // TODO: Implement edit user functionality when backend API is available
        console.log('Edit user functionality not yet implemented:', user);
        showSnackbar({
            type: 'info',
            title: 'Feature Not Available',
            message: 'Edit user functionality is not yet implemented'
        });
    };

    const handleDeleteUser = async (userId: string) => {
        // TODO: Implement delete user functionality when backend API is available
        console.log('Delete user functionality not yet implemented:', userId);
        showSnackbar({
            type: 'info',
            title: 'Feature Not Available',
            message: 'Delete user functionality is not yet implemented'
        });
    };

    return (
        <div className="space-y-6">
            {/* Organization Info */}
            {organizationId && (
                <div className="bg-card border border-border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                        <div>
                            <h2 className="text-lg font-semibold text-foreground">Organization</h2>
                            <p className="text-sm text-muted-foreground">ID: {organizationId}</p>
                            {currentUser && (
                                <p className="text-sm text-muted-foreground">
                                    Current User: {currentUser.email}
                                </p>
                            )}
                        </div>
                        <div className="text-right">
                            <p className="text-sm text-muted-foreground">
                                Total Users: {pagination.totalUsers}
                            </p>
                        </div>
                    </div>
                </div>
            )}

            {/* Header with Title and Create Buttons */}
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold text-foreground">Users</h1>
                <div className="flex space-x-2">
                    <Button
                        onClick={handleCreateOrganization}
                        disabled={loading || isCreatingOrg}
                        variant="outline"
                    >
                        {isCreatingOrg ? (
                            <>
                                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary mr-2"></div>
                                Creating Organization...
                            </>
                        ) : (
                            <>
                                <Building className="mr-2 h-4 w-4" />
                                Create Organization
                            </>
                        )}
                    </Button>
                    <CreateUserDialog
                        onUserCreated={handleCreateUser}
                        loading={isCreatingUser}
                    />
                </div>
            </div>



            {/* Users Table */}
            <div className="bg-card border border-border rounded-lg overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead className="bg-muted/50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    User
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Email
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Organization
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Status
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Created
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                            {loading ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-8 text-center">
                                        <div className="flex items-center justify-center">
                                            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                                            <span className="ml-2 text-muted-foreground">Loading users...</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : !users || users.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-8 text-center">
                                        <div className="flex flex-col items-center">
                                            <UserIcon className="h-12 w-12 text-muted-foreground mb-4" />
                                            <p className="text-muted-foreground">No users found</p>
                                            <p className="text-sm text-muted-foreground">Create an organization or add users to get started</p>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                users.map((user) => (
                                    <tr key={user.id} className="hover:bg-accent/50 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex items-center">
                                                <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center mr-3">
                                                    <UserIcon className="h-4 w-4 text-primary" />
                                                </div>
                                                <div>
                                                    <div className="text-sm font-medium text-foreground">
                                                        {user.username}
                                                    </div>
                                                    <div className="text-sm text-muted-foreground">
                                                        ID: {user.id}
                                                    </div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm text-foreground">{user.email}</div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm text-foreground">{user.organization_id}</div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${user.is_active
                                                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                                : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                                                }`}>
                                                {user.is_active ? 'Active' : 'Inactive'}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="text-sm text-foreground">
                                                {new Date(user.created_at).toLocaleDateString()}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                                            <div className="flex space-x-2">
                                                <button
                                                    className="text-primary hover:text-primary/80 transition-colors"
                                                    onClick={() => handleEditUser(user)}
                                                >
                                                    <Edit className="h-4 w-4" />
                                                </button>
                                                <button
                                                    className="text-destructive hover:text-destructive/80 transition-colors"
                                                    onClick={() => handleDeleteUser(user.id)}
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
