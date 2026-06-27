<script setup lang="ts">
import { computed, ref } from 'vue'

// prettier-ignore
const reasons = [
  { name: 'Total', total: true, color: 'bg-slate-500/80' },
  { name: 'BanInBlockedSubnet', color: 'bg-red-500/80', explanation: 'Ban peers in banned subnets' },
  { name: 'BanWeirdClient', color: 'bg-orange-500/80', explanation: 'Ban peers with weird client names' },
  { name: 'BanLeecherClient', color: 'bg-amber-500/80', explanation: 'Ban peers with well-known leecher client names' },
  { name: 'BanObsoleteClient', color: 'bg-yellow-500/80', explanation: 'Ban peers with obsolete client names' },
  { name: 'BanUploadedMoreThanTotalSize', color: 'bg-lime-500/80', explanation: 'Ban peers with uploaded data more than torrent total size' },
  { name: 'BanNoProgress', color: 'bg-green-500/80', explanation: 'Ban peers with uploaded data exceeding 10 MB and no progress' },
  { name: 'BanShrunkProgress', color: 'bg-emerald-500/80', explanation: 'Ban peers with shrunk progress' },
  { name: 'BanUploadedExcessively', color: 'bg-teal-500/80', explanation: 'Ban peers with uploaded data more than increased progress' },
  { name: 'BanSubnetTooManyPeersBanned', color: 'bg-cyan-500/80', explanation: 'Ban subnets with more than 4 banned peers' },
  { name: 'BanSubnetTooManyPeers', color: 'bg-sky-500/80', explanation: 'Ban subnets with more than 4 peers' },
]

type Log = { id: number; type: number; date: string; peer: string; client: string }

const apiVer = ref({ version: '', supported: false })
const stats = ref<Record<number, { session: number; all: number }>>({})
const logs = ref<Log[]>([])
const session = computed(() =>
  Object.values(stats.value).reduce((prev, curr) => prev + curr.session, 0),
)
const all = computed(() => Object.values(stats.value).reduce((prev, curr) => prev + curr.all, 0))
const hasMoreLogs = ref(true)
const loadingLogs = ref(false)

fetch('/api/apiVersion')
  .then(response => response.json())
  .then(json => (apiVer.value = json))
fetch('/api/stats')
  .then(response => response.json())
  .then(json => (stats.value = json))
fetch('/api/logs')
  .then(response => response.json())
  .then(json => (logs.value = json))

async function loadMoreLogs() {
  if (loadingLogs.value) return
  loadingLogs.value = true
  try {
    const id = logs.value.at(-1)?.id
    if (id == null) {
      hasMoreLogs.value = false
      return
    }
    const resp = await fetch(`/api/logs?before=${id}`)
    const json: Log[] = await resp.json()
    hasMoreLogs.value = json.length > 0
    logs.value = [...logs.value, ...json]
  } finally {
    loadingLogs.value = false
  }
}
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
        v-for="({ name, total, color, explanation }, i) in reasons"
        :key="name"
      >
        <div
          class="border-b border-current/30 p-1.5 text-center text-white"
          :class="color"
          :title="explanation"
        >
          {{ name }}
        </div>
        <div class="flex">
          <span class="flex flex-1 flex-col border-r border-current/30 text-center">
            <span class="p-1 pb-0 text-xs whitespace-nowrap text-current/80 select-none">
              This Session
            </span>
            <span class="px-2">{{ total ? session : (stats[i]?.session ?? 0) }}</span>
          </span>
          <span class="flex flex-1 flex-col text-center">
            <span class="p-1 pb-0 text-xs text-current/80 select-none">All</span>
            <span class="px-2">{{ total ? all : (stats[i]?.all ?? 0) }}</span>
          </span>
        </div>
      </div>
    </div>
  </div>

  <div class="p-4 text-sm">
    <table
      class="w-full border-separate border-spacing-0 overflow-hidden rounded-xl border border-current/30 text-current/80"
    >
      <thead class="text-center text-current/80 select-none">
        <tr>
          <th class="border-r border-current/30 p-2">Date</th>
          <th class="border-r border-current/30 p-2">Type</th>
          <th class="border-r border-current/30 p-2">Peer</th>
          <th class="p-2">Client</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="{ id, type, date, peer, client } in logs" :key="id" class="hover:bg-current/10">
          <td class="border-t border-r border-current/30 p-2 text-center">
            {{ new Date(date).toLocaleString() }}
          </td>
          <td class="border-t border-r border-current/30 p-2 text-center">
            <span
              class="rounded-full p-1 px-2.5 text-xs text-white"
              :class="reasons[type]?.color"
              :title="reasons[type]?.explanation"
            >
              {{ type === 0 ? 'ClearBannedIPs' : reasons[type]?.name }}
            </span>
          </td>
          <td class="border-t border-r border-current/30 p-2">{{ peer }}</td>
          <td class="border-t border-current/30 p-2">{{ client }}</td>
        </tr>
        <tr v-if="hasMoreLogs">
          <td
            class="border-t border-current/30 py-1 text-center select-none hover:bg-current/10"
            :class="{ 'hover:cursor-pointer': !loadingLogs }"
            colspan="4"
            @click="loadMoreLogs"
          >
            {{ loadingLogs ? 'Loading' : 'Load more' }}...
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style>
:root {
  color-scheme: light dark;
}
</style>
