import { useChannelMonitorFormat as rawFormat, providerGradient as rawGradient } from './useChannelMonitorFormat'
import { userPlatform } from '@/utils/userPlatformColors'

/** Display aliases only; never change monitor filters or provider IDs sent to the API. */
export function useChannelMonitorFormat() {
  const format = rawFormat()
  return {
    ...format,
    providerLabel: (p: string) => format.providerLabel(userPlatform(p)),
    providerBadgeClass: (p: string) => format.providerBadgeClass(userPlatform(p)),
    providerPickerClass: (p: string, active: boolean) => format.providerPickerClass(userPlatform(p), active)
  }
}

export const providerGradient = (p: string) => rawGradient(userPlatform(p))
export { hslForPct } from './useChannelMonitorFormat'
