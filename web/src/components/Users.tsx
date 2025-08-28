'use client';

import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { userActions } from '@/store/sagas/userSaga';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { User, Building } from 'lucide-react';
import { gauthApi, CreateOrganizationRequest } from '@/services/gauthApi';
import CreateUserDialog from './CreateUserDialog';

export default function Users() {
    const dispatch = useAppDispatch();
    const { users, loading, error } = useAppSelector((state) => state.user);
    const [isCreatingOrg, setIsCreatingOrg] = useState(false);
    const [isCreatingUser, setIsCreatingUser] = useState(false);
    const [orgError, setOrgError] = useState<string | null>(null);
    const [orgSuccess, setOrgSuccess] = useState<string | null>(null);
    const [userError, setUserError] = useState<string | null>(null);
    const [userSuccess, setUserSuccess] = useState<string | null>(null);

    useEffect(() => {
        dispatch(userActions.fetchUsers());
    }, [dispatch]);

    const handleCreateOrganization = async () => {
        try {
            setIsCreatingOrg(true);
            setOrgError(null);
            setOrgSuccess(null);

            // Create organization with initial user
            const orgData: CreateOrganizationRequest = {
                name: 'ODEYS Organization',
                initial_user_email: 'admin@odeys.com',
                initial_user_public_key: '0x1234567890abcdef1234567890abcdef12345678'
            };

            const response = await gauthApi.createOrganization(orgData);

            if (response.success) {
                setOrgSuccess(`Organization created successfully! Organization ID: ${response.data.organization.id}, User ID: ${response.data.user_id}`);

                // Refresh users list after creating organization
                dispatch(userActions.fetchUsers());
            } else {
                setOrgError('Failed to create organization');
            }
        } catch (error) {
            console.error('Error creating organization:', error);
            setOrgError(error instanceof Error ? error.message : 'Failed to create organization');
        } finally {
            setIsCreatingOrg(false);
        }
    };

    const handleCreateUser = async (userData: { name: string; email: string; role: string }) => {
        try {
            setIsCreatingUser(true);
            setUserError(null);
            setUserSuccess(null);

            // Call Redux action to create user
            dispatch(userActions.createUser({
                name: userData.name,
                email: userData.email,
                role: userData.role as 'admin' | 'user'
            }));

            setUserSuccess(`User "${userData.name}" created successfully!`);

            // Refresh users list after creating user
            dispatch(userActions.fetchUsers());
        } catch (error) {
            console.error('Error creating user:', error);
            setUserError(error instanceof Error ? error.message : 'Failed to create user');
        } finally {
            setIsCreatingUser(false);
        }
    };

    return (
        <div className="space-y-6">
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

            {/* Success Displays */}
            {orgSuccess && (
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg dark:bg-green-950 dark:border-green-800">
                    <div className="text-green-800 dark:text-green-200">
                        <h3 className="font-semibold">Success:</h3>
                        <p>{orgSuccess}</p>
                    </div>
                </div>
            )}

            {userSuccess && (
                <div className="p-4 bg-green-50 border border-green-200 rounded-lg dark:bg-green-950 dark:border-green-800">
                    <div className="text-green-800 dark:text-green-200">
                        <h3 className="font-semibold">Success:</h3>
                        <p>{userSuccess}</p>
                    </div>
                </div>
            )}

            {/* Error Display */}
            {(error || orgError || userError) && (
                <div className="p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
                    <div className="text-destructive">
                        <h3 className="font-semibold">Error:</h3>
                        <p>{userError || orgError || error}</p>
                    </div>
                </div>
            )}

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
                                    Role
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Status
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                            {loading ? (
                                <tr>
                                    <td colSpan={5} className="px-6 py-8 text-center">
                                        <div className="flex items-center justify-center">
                                            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
                                            <span className="ml-2 text-muted-foreground">Loading users...</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : users.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-6 py-8 text-center">
                                        <div className="flex flex-col items-center">
                                            <User className="h-12 w-12 text-muted-foreground mb-4" />
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
                                                    <User className="h-4 w-4 text-primary" />
                                                </div>
                                                <div>
                                                    <div className="text-sm font-medium text-foreground">
                                                        {user.name}
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
                                            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${user.role === 'admin'
                                                ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                                                : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                                }`}>
                                                {user.role}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                                                Active
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                                            <div className="flex space-x-2">
                                                <button className="text-primary hover:text-primary/80 transition-colors">
                                                    Edit
                                                </button>
                                                <button className="text-destructive hover:text-destructive/80 transition-colors">
                                                    Delete
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
