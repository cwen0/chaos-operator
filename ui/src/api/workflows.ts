import { Workflow, WorkflowSingle, workflowParams } from './workflows.type'

import http from './http'

export const newWorkflow = (data: any) => http.post('/workflows', data)

export const workflows = (params?: workflowParams) =>
  http.get<Workflow[]>('/workflows', {
    params,
  })

export const detail = (ns: string, name: string) => http.get<WorkflowSingle>(`/workflows/${ns}/${name}`)

export const update = (ns: string, name: string, data: WorkflowSingle['kube_object']) =>
  http.put(`/workflows/${ns}/${name}`, data)

export const del = (ns: string, name: string) => http.delete(`/workflows/${ns}/${name}`)
