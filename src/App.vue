<script setup lang="ts">
import { computed, onErrorCaptured, ref } from 'vue'

import Logs from './Logs.vue'
import { reasons, xfetch } from './utils.ts'

const apiVer = ref({ version: '', supported: false })
const stats = ref<Record<number, { session: number; all: number }>>({})
const session = computed(() =>
  Object.values(stats.value).reduce((prev, curr) => prev + curr.session, 0),
)
const all = computed(() => Object.values(stats.value).reduce((prev, curr) => prev + curr.all, 0))

const loading = ref(true)
const errors = ref(new Map<number, string>())
const onError = (error: any) => {
  const key = Math.random()
  errors.value.set(key, String(error))
  setTimeout(() => errors.value.delete(key), 3_000)
}
onErrorCaptured(onError)
xfetch('/api/apiVersion').then(json => (apiVer.value = json), onError)
xfetch('/api/stats').then(json => ([stats.value, loading.value] = [json, false]), onError)
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
            <span class="px-2">
              {{ loading ? '-' : total ? session : (stats[i]?.session ?? 0) }}
            </span>
          </span>
          <span class="flex flex-1 flex-col text-center hover:bg-current/10">
            <span class="p-1 pb-0 text-xs text-current/80 select-none">All</span>
            <span class="px-2">{{ loading ? '-' : total ? all : (stats[i]?.all ?? 0) }}</span>
          </span>
        </div>
      </div>
    </div>
  </div>
  <Suspense>
    <Logs />
  </Suspense>

  <div class="fixed right-4 bottom-4 space-y-3">
    <div
      v-for="[key, error] in errors"
      :key="key"
      class="rounded-lg bg-red-700/80 px-3 py-2 text-white shadow-xl"
    >
      <p class="text-sm">{{ error }}</p>
    </div>
  </div>
</template>

<style>
@import './style.css';
</style>
