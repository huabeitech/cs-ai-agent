import { fetchTicketPriorityConfigsAll } from "@/lib/api/ticket-config"

export type TicketPriorityOption = {
  value: string
  label: string
}

export async function getTicketPriorityOptions() {
  const list = await fetchTicketPriorityConfigsAll()
  return (Array.isArray(list) ? list : []).map((item) => ({
    value: String(item.id),
    label: item.name,
  })) satisfies TicketPriorityOption[]
}

export async function getTicketPriorityMap() {
  const options = await getTicketPriorityOptions()
  return Object.fromEntries(options.map((item) => [Number(item.value), item.label])) as Record<number, string>
}
