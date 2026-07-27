<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'

const reasons = [
  {
    name: 'Total',
    total: true,
    color: 'text-slate-500/80',
    bg: 'bg-slate-500/80',
    bgLighter: 'bg-slate-500/20',
  },
  {
    name: 'BanInBlockedSubnet',
    color: 'text-red-500/80',
    bg: 'bg-red-500/80',
    bgLighter: 'bg-red-500/20',
    explanation: 'Ban peers in banned subnets',
  },
  {
    name: 'BanWeirdClient',
    color: 'text-orange-500/80',
    bg: 'bg-orange-500/80',
    bgLighter: 'bg-orange-500/20',
    explanation: 'Ban peers with weird client names',
  },
  {
    name: 'BanLeecherClient',
    color: 'text-amber-500/80',
    bg: 'bg-amber-500/80',
    bgLighter: 'bg-amber-500/20',
    explanation: 'Ban peers with well-known leecher client names',
  },
  {
    name: 'BanObsoleteClient',
    color: 'text-yellow-500/80',
    bg: 'bg-yellow-500/80',
    bgLighter: 'bg-yellow-500/20',
    explanation: 'Ban peers with obsolete client names',
  },
  {
    name: 'BanUploadedMoreThanTotalSize',
    color: 'text-lime-500/80',
    bg: 'bg-lime-500/80',
    bgLighter: 'bg-lime-500/20',
    explanation: 'Ban peers with uploaded data more than torrent total size',
  },
  {
    name: 'BanNoProgress',
    color: 'text-green-500/80',
    bg: 'bg-green-500/80',
    bgLighter: 'bg-green-500/20',
    explanation: 'Ban peers with uploaded data exceeding 10 MB and no progress',
  },
  {
    name: 'BanShrunkProgress',
    color: 'text-emerald-500/80',
    bg: 'bg-emerald-500/80',
    bgLighter: 'bg-emerald-500/20',
    explanation: 'Ban peers with shrunk progress',
  },
  {
    name: 'BanUploadedExcessively',
    color: 'text-teal-500/80',
    bg: 'bg-teal-500/80',
    bgLighter: 'bg-teal-500/20',
    explanation: 'Ban peers with uploaded data more than increased progress',
  },
  {
    name: 'BanSubnetTooManyPeersBanned',
    color: 'text-cyan-500/80',
    bg: 'bg-cyan-500/80',
    bgLighter: 'bg-cyan-500/20',
    explanation: 'Ban subnets with more than 4 banned peers',
  },
  {
    name: 'BanSubnetTooManyPeers',
    color: 'text-sky-500/80',
    bg: 'bg-sky-500/80',
    bgLighter: 'bg-sky-500/20',
    explanation: 'Ban subnets with more than 4 peers',
  },
]

type Log = { d: string; t: number; p?: string; c?: string }
const peerPat = /(.+?)(:\d+)?$/

const apiVer = ref({ version: '', supported: false })
const stats = ref<Record<number, { session: number; all: number }>>({})
const logs = ref<Log[]>([])
const session = computed(() =>
  Object.values(stats.value).reduce((prev, curr) => prev + curr.session, 0),
)
const all = computed(() => Object.values(stats.value).reduce((prev, curr) => prev + curr.all, 0))
const logGroups = computed(() => {
  const groups = new Map<string, Log[]>()
  for (const log of logs.value) {
    const date = new Date(log.d).toLocaleDateString()
    let group = groups.get(date)
    if (!group) groups.set(date, (group = []))
    group.push(log)
  }
  return groups
})

const sentry = useTemplateRef('sentry')
const isIntersecting = ref(false)
const hasMore = ref(false)
const loading = ref(false)
let observer: IntersectionObserver | null = null
onMounted(
  () =>
    (observer = new IntersectionObserver(
      ([entry]) => (isIntersecting.value = entry?.isIntersecting ?? false),
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
    const date = logs.value.at(-1)?.d
    if (date == null) {
      loading.value = false
      return
    }
    const resp = await fetch(`/api/logs?before=${date}`)
    const json: Log[] = await resp.json()
    hasMore.value = json.length > 0
    logs.value = [...logs.value, ...json]
    loading.value = false
  }
})

fetch('/api/apiVersion')
  .then(response => response.json())
  .then(json => (apiVer.value = json))
fetch('/api/stats')
  .then(response => response.json())
  .then(json => (stats.value = json))
fetch('/api/logs')
  .then(response => response.json())
  .then(json => {
    logs.value = json
    hasMore.value = logs.value.length > 0
  })
</script>

<template>
  <div class="m-4 rounded-xl border border-current/30 p-3 text-sm text-current/80">
    <div class="whitespace-nowrap select-none">
      qBittorrent Web API Version:
      <span
        class="ml-1 rounded-full p-1 px-1.5 text-xs text-white outline outline-offset-1"
        :class="[
          apiVer.version
            ? apiVer.supported
              ? 'bg-green-500/80 outline-green-500/80'
              : 'bg-red-500/85 outline-red-500/85'
            : 'bg-gray-500/80 outline-gray-500/80',
        ]"
        :title="apiVer.version && apiVer.supported ? undefined : 'qBittorrent API >= v2.3 required'"
      >
        {{ apiVer.version ? `v${apiVer.version}` : 'Unknown' }}
      </span>
    </div>

    <div class="mt-2">
      <div
        class="m-1 inline-block w-fit overflow-clip rounded-xl border border-current/30"
        v-for="({ name, total, bg, explanation }, i) in reasons"
        :key="name"
      >
        <div
          class="border-b border-current/30 p-1.5 text-center text-white"
          :class="bg"
          :title="explanation"
        >
          {{ name }}
        </div>
        <div class="flex">
          <span
            class="flex flex-1 flex-col border-r border-current/30 text-center hover:bg-current/10"
          >
            <span class="p-1 pb-0 text-xs whitespace-nowrap text-current/80 select-none">
              This Session
            </span>
            <span class="px-2">{{ total ? session : (stats[i]?.session ?? 0) }}</span>
          </span>
          <span class="flex flex-1 flex-col text-center hover:bg-current/10">
            <span class="p-1 pb-0 text-xs text-current/80 select-none">All</span>
            <span class="px-2">{{ total ? all : (stats[i]?.all ?? 0) }}</span>
          </span>
        </div>
      </div>
    </div>
  </div>

  <div v-for="[date, group] in logGroups" :key="date">
    <div class="sticky top-0 bg-[canvas] pt-1 text-center">
      <span class="text-current/70">{{ date }}</span>
      <div class="mt-1 h-px bg-linear-to-r from-current/0 via-current/30 to-current/0"></div>
    </div>
    <div class="m-3 text-sm">
      <div
        class="flex items-center gap-1.5 rounded-lg from-current/6 via-current/6 to-current/0 p-0.75 px-1.5 whitespace-nowrap hover:bg-linear-to-r"
        v-for="{ d: date, t: type, p: peer, c: client } in group"
        :key="date"
      >
        <span class="text-current/70">
          {{ new Date(date).toLocaleTimeString() }}
        </span>
        <span
          class="rounded-lg p-1 px-2 text-xs font-bold"
          :class="`${reasons[type]?.color} ${reasons[type]?.bgLighter}`"
          :title="reasons[type]?.explanation"
        >
          {{ type === 0 ? 'ClearBannedIPs' : reasons[type]?.name }}
        </span>
        <span v-if="peer">
          <span class="text-current/50">peer=</span>
          <span class="text-current/80">{{ peer.match(peerPat)?.[1] }}</span>
          <span class="text-current/50">{{ peer.match(peerPat)?.[2] ?? '' }}</span>
        </span>
        <span v-if="client">
          <span class="text-current/50">client=</span>
          <span class="text-current/80">{{ client }}</span>
        </span>
      </div>
    </div>
  </div>
  <div
    v-if="loading || hasMore"
    class="mb-3 h-4 w-4 animate-spin justify-self-center rounded-full border-3 border-current/20 border-t-current/70"
    ref="sentry"
  ></div>
</template>

<style>
:root {
  color-scheme: light dark;
}
@media (orientation: landscape) {
  body {
    justify-self: center;
    width: 100vh;
  }
}
</style>
