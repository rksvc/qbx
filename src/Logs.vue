<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'

import { formatBytes, reasons, xfetch } from './utils'

type Log = { date: string; t: number; p?: string; c?: string; d?: number; u?: number }
const peerPat = /(.+?)(:\d+)?$/

const logs = ref<Log[]>(await xfetch('/api/logs'))
const hasMore = ref(logs.value.length > 0)
const logGroups = computed(() => {
  const groups = new Map<string, Log[]>()
  for (const log of logs.value) {
    const date = new Date(log.date).toLocaleDateString()
    let group = groups.get(date)
    if (!group) groups.set(date, (group = []))
    group.push(log)
  }
  return groups
})

const sentry = useTemplateRef('sentry')
const isIntersecting = ref(false)
const loading = ref(false)
let observer: IntersectionObserver | null = null
onMounted(
  () =>
    (observer = new IntersectionObserver(
      ([entry]) => (isIntersecting.value = entry?.isIntersecting ?? false),
      { rootMargin: '200px' },
    )),
)
onUnmounted(() => observer?.disconnect())
watch(sentry, (value, oldValue) => {
  if (oldValue) observer?.unobserve(oldValue)
  if (value) observer?.observe(value)
})
watch([isIntersecting, hasMore], async () => {
  if (isIntersecting.value && !loading.value && hasMore.value) {
    loading.value = true
    const date = logs.value.at(-1)?.date
    if (date == null) {
      loading.value = false
      return
    }
    const resp: Log[] = await xfetch(`/api/logs?before=${date}`)
    hasMore.value = resp.length > 0
    logs.value = [...logs.value, ...resp]
    loading.value = false
  }
})
</script>

<template>
  <div v-for="[date, group] in logGroups" :key="date">
    <div class="sticky top-0 bg-[canvas] pt-1 text-center">
      <span class="text-current/70">{{ date }}</span>
      <div class="mt-1 h-px bg-linear-to-r from-current/0 via-current/30 to-current/0"></div>
    </div>
    <div class="m-4 mt-2 text-sm">
      <div
        v-for="{ date, t: type, p: peer, c: client, d: downloaded, u: uploaded } in group"
        :key="date"
        class="flex items-center gap-1.5 p-0.75 px-1.5 whitespace-nowrap"
      >
        <span class="text-current/70">
          {{ new Date(date).toLocaleTimeString() }}
        </span>
        <span
          class="rounded-full p-0.75 px-2 text-xs font-bold"
          :class="`${reasons[type]?.color} ${reasons[type]?.bgLighter}`"
          :title="reasons[type]?.explanation"
        >
          {{ type === 0 ? 'ClearBannedIPs' : reasons[type]?.name }}
        </span>
        <template v-if="type > 0 && type < 9">
          <span class="text-xs text-green-500">↑ {{ formatBytes(uploaded ?? 0) }}</span>
          <span class="text-xs text-red-500">↓ {{ formatBytes(downloaded ?? 0) }}</span>
        </template>
        <span v-if="peer">
          <span class="text-current/50 italic">peer=</span>
          <code class="rounded-md bg-current/7 p-0.5 px-1">
            <span class="text-current/80">{{ peer.match(peerPat)?.[1] }}</span>
            <span class="text-current/50">{{ peer.match(peerPat)?.[2] ?? '' }}</span>
          </code>
        </span>
        <span v-if="client">
          <span class="text-current/50 italic">client=</span>
          <code class="rounded-md bg-current/7 p-0.5 px-1 text-current/80">{{ client }}</code>
        </span>
      </div>
    </div>
  </div>
  <div
    v-if="loading || hasMore"
    class="mb-5 h-4 w-4 animate-spin justify-self-center rounded-full border-3 border-current/20 border-t-current/70"
    ref="sentry"
  ></div>
</template>
