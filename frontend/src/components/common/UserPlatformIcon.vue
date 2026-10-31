<template>
  <PlatformIcon :platform="userPlatform(platform)" :size="size" />
</template>

<script setup lang="ts">
import type { GroupPlatform } from '@/types'
import PlatformIcon from './PlatformIcon.vue'
import { userPlatform } from '@/utils/userPlatformColors'

defineProps<{ platform?: GroupPlatform; size?: 'xs' | 'sm' | 'md' | 'lg' }>()
</script>
