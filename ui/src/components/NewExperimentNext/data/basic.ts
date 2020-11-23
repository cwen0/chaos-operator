import * as Yup from 'yup'

const data = {
  name: '',
  namespace: '',
  labels: [],
  annotations: [],
  scope: {
    namespace_selectors: [],
    label_selectors: [],
    annotation_selectors: [],
    phase_selectors: ['all'],
    mode: 'one',
    value: '',
    pods: [],
  },
  scheduler: {
    cron: '',
    duration: '',
  },
}

export const schema: Yup.ObjectSchema = Yup.object({
  name: Yup.string().required('The experiment name is required'),
  namespace: Yup.string().required('The experiment namespace is required'),
  scheduler: Yup.object({
    cron: Yup.string().required('The cron is required'),
  }),
})

export type dataType = typeof data

export default data
