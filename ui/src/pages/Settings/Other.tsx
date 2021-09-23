import { MenuItem, TextField, Typography } from '@material-ui/core'
import { setLang, setTheme } from 'slices/settings'
import { useStoreDispatch, useStoreSelector } from 'store'

import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import messages from 'i18n/messages'

const Other = () => {
  const { settings } = useStoreSelector((state) => state)
  const { theme, lang } = settings
  const dispatch = useStoreDispatch()

  const handleChangeTheme = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setTheme(e.target.value))
  const handleChangeLang = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setLang(e.target.value))

  return (
    <>
      <PaperTop title={T('common.other')} divider />
      <Space>
        <TextField
          select
          size="small"
          style={{ width: '50%' }}
          label={T('settings.theme.title')}
          helperText={T('settings.theme.choose')}
          value={theme}
          onChange={handleChangeTheme}
        >
          <MenuItem value="light">
            <Typography>{T('settings.theme.light')}</Typography>
          </MenuItem>
          <MenuItem value="dark">
            <Typography>{T('settings.theme.dark')}</Typography>
          </MenuItem>
        </TextField>

        <TextField
          select
          size="small"
          style={{ width: '50%' }}
          label={T('settings.lang.title')}
          helperText={T('settings.lang.choose')}
          value={lang}
          onChange={handleChangeLang}
        >
          {Object.keys(messages).map((lang) => (
            <MenuItem key={lang} value={lang}>
              <Typography>{T(`settings.lang.${lang}`)}</Typography>
            </MenuItem>
          ))}
        </TextField>
      </Space>
    </>
  )
}

export default Other
