/**
 * Admin support ticket endpoints
 */

import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'
import type { Ticket, TicketMessage, TicketStatus } from '../tickets'

export interface AdminTicket extends Ticket {
  user_id: number
  user_email: string
}

export interface AdminTicketDetail extends AdminTicket {
  messages: TicketMessage[]
}

export async function list(
  page = 1,
  pageSize = 20,
  status: TicketStatus | '' = ''
): Promise<BasePaginationResponse<AdminTicket>> {
  const { data } = await apiClient.get<BasePaginationResponse<AdminTicket>>('/admin/tickets', {
    params: { page, page_size: pageSize, ...(status ? { status } : {}) }
  })
  return data
}

export async function getById(id: number): Promise<AdminTicketDetail> {
  const { data } = await apiClient.get<AdminTicketDetail>(`/admin/tickets/${id}`)
  return data
}

export async function reply(id: number, content: string): Promise<TicketMessage> {
  const { data } = await apiClient.post<TicketMessage>(`/admin/tickets/${id}/messages`, { content })
  return data
}

export async function updateStatus(id: number, status: TicketStatus): Promise<void> {
  await apiClient.put(`/admin/tickets/${id}/status`, { status })
}

const ticketsAPI = { list, getById, reply, updateStatus }

export default ticketsAPI
