import { cn } from '@/lib/utils'

describe('cn utility function', () => {
    it('merges class names correctly', () => {
        const result = cn('px-4', 'py-2', 'bg-blue-500')
        expect(result).toBe('px-4 py-2 bg-blue-500')
    })

    it('handles conditional classes', () => {
        const isActive = true
        const result = cn('base-class', isActive && 'active-class')
        expect(result).toBe('base-class active-class')
    })

    it('handles false conditional classes', () => {
        const isActive = false
        const result = cn('base-class', isActive && 'active-class')
        expect(result).toBe('base-class')
    })

    it('handles undefined and null values', () => {
        const result = cn('base-class', undefined, null, 'valid-class')
        expect(result).toBe('base-class valid-class')
    })

    it('handles empty strings', () => {
        const result = cn('base-class', '', 'valid-class')
        expect(result).toBe('base-class valid-class')
    })

    it('handles arrays of classes', () => {
        const result = cn(['class1', 'class2'], 'class3')
        expect(result).toBe('class1 class2 class3')
    })

    it('handles objects with boolean values', () => {
        const result = cn({
            'active': true,
            'disabled': false,
            'highlighted': true
        })
        expect(result).toBe('active highlighted')
    })

    it('merges conflicting Tailwind classes correctly', () => {
        // This tests the twMerge functionality
        const result = cn('px-4', 'px-6')
        expect(result).toBe('px-6') // px-6 should override px-4
    })

    it('handles complex combinations', () => {
        const isActive = true
        const isDisabled = false
        const result = cn(
            'base-class',
            isActive && 'active-class',
            isDisabled && 'disabled-class',
            ['array-class1', 'array-class2'],
            {
                'object-class': true,
                'false-class': false
            }
        )
        expect(result).toBe('base-class active-class array-class1 array-class2 object-class')
    })

    it('handles no arguments', () => {
        const result = cn()
        expect(result).toBe('')
    })

    it('handles single argument', () => {
        const result = cn('single-class')
        expect(result).toBe('single-class')
    })
})
