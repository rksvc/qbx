export const reasons = [
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

export async function xfetch(url: string) {
  const resp = await fetch(url)
  if (!resp.ok)
    if (resp.status === 500) throw new Error(await resp.json())
    else throw new Error(`${resp.status} ${resp.statusText}`)
  return await resp.json()
}

export function formatBytes(n: number) {
  return n < 1024
    ? `${n} B`
    : n < Math.pow(1024, 2)
      ? `${(n / 1024).toFixed(2)} KB`
      : n < Math.pow(1024, 3)
        ? `${(n / Math.pow(1024, 2)).toFixed(2)} MB`
        : `${(n / Math.pow(1024, 3)).toFixed(2)} GB`
}
