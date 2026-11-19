/**
 * Support ticket endpoints for the signed-in user
 */

import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'

export type TicketStatus = 'open' | 'answered' | 'closed'
export type TicketCategory = 'account' | 'billing' | 'usage' | 'other'
export type TicketAuthor = 'user' | 'staff' | 'ai'

export const TICKET_CATEGORIES: TicketCategory[] = ['usage', 'billing', 'account', 'other']
export const TICKET_SUBJECT_MAX = 200
export const TICKET_CONTENT_MAX = 5000

export interface Ticket {
  id: number
  subject: string
  category: TicketCategory
  status: TicketStatus
  last_message_at: string
  created_at: string
}

export interface TicketMessage {
  id: number
  author_type: TicketAuthor
  content: string
  created_at: string
}

export interface TicketDetail extends Ticket {
  messages: TicketMessage[]
}

export interface CreateTicketRequest {
  subject: string
  category: TicketCategory
  content: string
}

export async function list(page = 1, pageSize = 20): Promise<BasePaginationResponse<Ticket>> {
  const { data } = await apiClient.get<BasePaginationResponse<Ticket>>('/tickets', {
    params: { page, page_size: pageSize }
  })
  return data
}

export async function create(request: CreateTicketRequest): Promise<Ticket> {
  const { data } = await apiClient.post<Ticket>('/tickets', request)
  return data
}

export async function getById(id: number): Promise<TicketDetail> {
  const { data } = await apiClient.get<TicketDetail>(`/tickets/${id}`)
  return data
}

export async function reply(id: number, content: string): Promise<TicketMessage> {
  const { data } = await apiClient.post<TicketMessage>(`/tickets/${id}/messages`, { content })
  return data
}

export async function close(id: number): Promise<void> {
  await apiClient.post(`/tickets/${id}/close`)
}

const ticketsAPI = { list, create, getById, reply, close }

export default ticketsAPI
