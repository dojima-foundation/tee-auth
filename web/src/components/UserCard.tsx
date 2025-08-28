'use client';

import React from 'react';

interface User {
    id: number;
    name: string;
    email: string;
    role?: string;
    isActive?: boolean;
}

interface UserCardProps {
    user: User;
    onEdit?: (user: User) => void;
    onDelete?: (userId: number) => void;
}

export const UserCard: React.FC<UserCardProps> = ({ user, onEdit, onDelete }) => {
    return (
        <div
            className="bg-white shadow-md rounded-lg p-6 border"
            data-testid={`user-card-${user.id}`}
        >
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900" data-testid="user-name">
                    {user.name}
                </h3>
                <div className="flex items-center space-x-2">
                    {user.isActive ? (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                            Active
                        </span>
                    ) : (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                            Inactive
                        </span>
                    )}
                </div>
            </div>

            <div className="space-y-2">
                <p className="text-sm text-gray-600" data-testid="user-email">
                    <span className="font-medium">Email:</span> {user.email}
                </p>
                {user.role && (
                    <p className="text-sm text-gray-600" data-testid="user-role">
                        <span className="font-medium">Role:</span> {user.role}
                    </p>
                )}
            </div>

            <div className="mt-4 flex space-x-2">
                {onEdit && (
                    <button
                        onClick={() => onEdit(user)}
                        className="px-3 py-1 text-sm font-medium text-blue-600 bg-blue-100 rounded-md hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
                        data-testid="edit-user-button"
                    >
                        Edit
                    </button>
                )}
                {onDelete && (
                    <button
                        onClick={() => onDelete(user.id)}
                        className="px-3 py-1 text-sm font-medium text-red-600 bg-red-100 rounded-md hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-red-500"
                        data-testid="delete-user-button"
                    >
                        Delete
                    </button>
                )}
            </div>
        </div>
    );
};
