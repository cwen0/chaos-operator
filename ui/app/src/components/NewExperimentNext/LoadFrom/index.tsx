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

import { Box, Divider, FormControlLabel, Radio, RadioGroup, Typography } from '@mui/material'
import {
  PkgDashboardApiserverArchiveArchive,
  PkgDashboardApiserverExperimentExperiment,
  PkgDashboardApiserverScheduleSchedule,
} from 'openapi'
import { PreDefinedValue, getDB } from 'lib/idb'
import { useEffect, useState } from 'react'

import Paper from '@ui/mui-extends/esm/Paper'
import RadioLabel from './RadioLabel'
import SkeletonN from '@ui/mui-extends/esm/SkeletonN'
import Space from '@ui/mui-extends/esm/Space'
import T from 'components/T'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

interface LoadFromProps {
  callback?: (data: any) => void
  inSchedule?: boolean
  inWorkflow?: boolean
}

const LoadFrom: React.FC<LoadFromProps> = ({ callback, inSchedule, inWorkflow }) => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<{
    experiments: PkgDashboardApiserverExperimentExperiment[]
    archives: PkgDashboardApiserverArchiveArchive[]
    schedules: PkgDashboardApiserverScheduleSchedule[]
  }>({
    experiments: [],
    archives: [],
    schedules: [],
  })
  const [predefined, setPredefined] = useState<PreDefinedValue[]>([])
  const [radio, setRadio] = useState('')

  useEffect(() => {
    const fetchExperiments = api.experiments.experimentsGet
    const fetchArchives = inSchedule ? api.archives.archivesSchedulesGet : api.archives.archivesGet
    const promises: Promise<any>[] = [fetchExperiments(), fetchArchives()]

    if (inSchedule) {
      promises.push(api.schedules.schedulesGet())
    }

    const fetchAll = async () => {
      try {
        const data = await Promise.all(promises)

        setData({
          experiments: data[0].data,
          archives: data[1].data,
          schedules: data[2] ? data[2].data : [],
        })
      } catch (error) {
        console.error(error)
      }

      let _predefined = await (await getDB()).getAll('predefined')
      if (!inSchedule) {
        _predefined = _predefined.filter((d) => d.kind !== 'Schedule')
      }
      setPredefined(_predefined)

      setLoading(false)
    }

    fetchAll()
  }, [inSchedule, inWorkflow])

  const onRadioChange = (e: any) => {
    const [type, uuid] = e.target.value.split('+')

    if (type === 'p') {
      const experiment = predefined?.filter((p) => p.name === uuid)[0].yaml

      callback && callback(experiment)

      dispatch(
        setAlert({
          type: 'success',
          message: T('confirm.success.load', intl),
        })
      )

      return
    }

    let apiRequest
    switch (type) {
      case 's':
        apiRequest = api.schedules.schedulesUidGet
        break
      case 'e':
        apiRequest = api.experiments.experimentsUidGet
        break
      case 'a':
        apiRequest = inSchedule ? api.archives.archivesSchedulesUidGet : api.archives.archivesUidGet
        break
    }

    setRadio(e.target.value)

    if (apiRequest) {
      apiRequest({ uid: uuid })
        .then(({ data }) => {
          callback && callback(data.kube_object)

          dispatch(
            setAlert({
              type: 'success',
              message: T('confirm.success.load', intl),
            })
          )
        })
        .catch(console.error)
    }
  }

  return (
    <Paper>
      <RadioGroup value={radio} onChange={onRadioChange}>
        <Space>
          {inSchedule && (
            <>
              <Typography>{T('schedules.title')}</Typography>

              {loading ? (
                <SkeletonN n={3} />
              ) : data.schedules.length > 0 ? (
                <Box display="flex" flexWrap="wrap">
                  {data.schedules.map((d) => (
                    <FormControlLabel
                      key={d.uid}
                      value={`s+${d.uid}`}
                      control={<Radio color="primary" />}
                      label={RadioLabel(d.name!, d.uid)}
                    />
                  ))}
                </Box>
              ) : (
                <Typography variant="body2" color="textSecondary">
                  {T('schedules.notFound')}
                </Typography>
              )}

              <Divider />
            </>
          )}

          {!inSchedule && (
            <>
              <Typography>{T('experiments.title')}</Typography>

              {loading ? (
                <SkeletonN n={3} />
              ) : data.experiments.length > 0 ? (
                <Box display="flex" flexWrap="wrap">
                  {data.experiments.map((d) => (
                    <FormControlLabel
                      key={d.uid}
                      value={`e+${d.uid}`}
                      control={<Radio color="primary" />}
                      label={RadioLabel(d.name!, d.uid)}
                    />
                  ))}
                </Box>
              ) : (
                <Typography variant="body2" color="textSecondary">
                  {T('experiments.notFound')}
                </Typography>
              )}

              <Divider />
            </>
          )}

          <Typography>{T('archives.title')}</Typography>

          {loading ? (
            <SkeletonN n={3} />
          ) : data.archives.length > 0 ? (
            <Box display="flex" flexWrap="wrap">
              {data.archives.map((d) => (
                <FormControlLabel
                  key={d.uid}
                  value={`a+${d.uid}`}
                  control={<Radio color="primary" />}
                  label={RadioLabel(d.name!, d.uid)}
                />
              ))}
            </Box>
          ) : (
            <Typography variant="body2" color="textSecondary">
              {T('archives.notFound')}
            </Typography>
          )}

          <Divider />

          <Typography>{T('dashboard.predefined')}</Typography>

          {loading ? (
            <SkeletonN n={3} />
          ) : predefined.length > 0 ? (
            <Box display="flex" flexWrap="wrap">
              {predefined.map((d) => (
                <FormControlLabel
                  key={d.name}
                  value={`p+${d.name}`}
                  control={<Radio color="primary" />}
                  label={RadioLabel(d.name)}
                />
              ))}
            </Box>
          ) : (
            <Typography variant="body2" color="textSecondary">
              {T('dashboard.noPredefinedFound')}
            </Typography>
          )}
        </Space>
      </RadioGroup>
    </Paper>
  )
}

export default LoadFrom
