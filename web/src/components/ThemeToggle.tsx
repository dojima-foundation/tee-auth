'use client';

import { Moon, Sun, Monitor } from 'lucide-react';
import { useAppDispatch, useAppSelector } from '@/store/hooks';
import { themeActions } from '@/store/sagas/themeSaga';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

export function ThemeToggle() {
    const dispatch = useAppDispatch();
    const { currentTheme, loading } = useAppSelector((state) => state.theme);

    const handleThemeChange = (theme: 'light' | 'dark' | 'system') => {
        dispatch(themeActions.saveTheme(theme));
    };

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <button
                    className="relative p-2 hover:bg-accent rounded-lg transition-colors duration-200 flex items-center justify-center"
                    disabled={loading}
                >
                    <Sun className="h-5 w-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
                    <Moon className="absolute h-5 w-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
                    <span className="sr-only">Toggle theme</span>
                </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <DropdownMenuItem
                    onClick={() => handleThemeChange('light')}
                    className={currentTheme === 'light' ? 'bg-accent' : ''}
                >
                    <Sun className="mr-2 h-4 w-4" />
                    <span>Light</span>
                </DropdownMenuItem>
                <DropdownMenuItem
                    onClick={() => handleThemeChange('dark')}
                    className={currentTheme === 'dark' ? 'bg-accent' : ''}
                >
                    <Moon className="mr-2 h-4 w-4" />
                    <span>Dark</span>
                </DropdownMenuItem>
                <DropdownMenuItem
                    onClick={() => handleThemeChange('system')}
                    className={currentTheme === 'system' ? 'bg-accent' : ''}
                >
                    <Monitor className="mr-2 h-4 w-4" />
                    <span>System</span>
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
