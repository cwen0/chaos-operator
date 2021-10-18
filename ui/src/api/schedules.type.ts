/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { Experiment } from './experiments.type'

export interface ScheduleParams {
  namespace?: string
}

export type Schedule = { is: 'schedule' } & Omit<Experiment, 'is'>

export interface ScheduleSingle extends Schedule {
  experiment_uids: uuid[]
  kube_object: any
}
