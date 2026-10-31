/** User-facing branding only. Never use these aliases for routing or account management. */
import * as raw from './platformColors'

export function userPlatform<T extends string | undefined>(platform: T): T | 'deepseek' {
  return platform === 'opencode_go' ? 'deepseek' : platform
}

export function userGroupName(name: string): string {
  return name.replace(/opencode(?:[\s_-]+(?:go|zen))?/gi, 'DeepSeek')
}

export const platformLabel = (p: string) => raw.platformLabel(userPlatform(p))
export const platformBadgeClass = (p: string) => raw.platformBadgeClass(userPlatform(p))
export const platformBadgeLightClass = (p: string) => raw.platformBadgeLightClass(userPlatform(p))
export const platformBorderClass = (p: string) => raw.platformBorderClass(userPlatform(p))
export const platformBorderStrongClass = (p: string) => raw.platformBorderStrongClass(userPlatform(p))
export const platformAccentColor = (p: string) => raw.platformAccentColor(userPlatform(p))
export const platformAccentBarClass = (p: string) => raw.platformAccentBarClass(userPlatform(p))
export const platformTextClass = (p: string) => raw.platformTextClass(userPlatform(p))
export const platformIconClass = (p: string) => raw.platformIconClass(userPlatform(p))
export const platformButtonClass = (p: string) => raw.platformButtonClass(userPlatform(p))
export const platformDiscountClass = (p: string) => raw.platformDiscountClass(userPlatform(p))
export const platformGradientClass = (p: string) => raw.platformGradientClass(userPlatform(p))
export const platformGradientTextClass = (p: string) => raw.platformGradientTextClass(userPlatform(p))
export const platformGradientSubtextClass = (p: string) => raw.platformGradientSubtextClass(userPlatform(p))
