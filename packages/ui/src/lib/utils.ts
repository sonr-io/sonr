import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Utility function for merging Tailwind CSS classes with proper conflict resolution.
 *
 * This function combines clsx for conditional class names with tailwind-merge
 * for handling Tailwind CSS class conflicts. It ensures that conflicting
 * utility classes are properly resolved (e.g., "px-4 px-6" becomes "px-6").
 *
 * @param inputs - Class values that can be strings, objects, arrays, or conditional expressions
 * @returns A string of merged and deduplicated CSS classes
 *
 * @example
 * ```ts
 * cn("px-4 py-2", "bg-blue-500", { "text-white": true, "text-black": false })
 * // Returns: "px-4 py-2 bg-blue-500 text-white"
 *
 * cn("px-4", "px-6") // Conflicting classes - returns: "px-6"
 * cn("bg-red-500", condition && "bg-blue-500") // Conditional classes
 * ```
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
