<template>
  <GroupBadge v-bind="$attrs" :name="userGroupName(name)" :platform="userPlatform(platform)" />
</template>

<script setup lang="ts">
import type { GroupPlatform } from '@/types'
import GroupBadge from './GroupBadge.vue'
import { userGroupName, userPlatform } from '@/utils/userPlatformColors'

defineOptions({ inheritAttrs: false })
defineProps<{ name: string; platform?: GroupPlatform }>()
</script>
