import { AppBar, IconButton, Toolbar } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import ArchiveOutlinedIcon from '@material-ui/icons/ArchiveOutlined'
import BlurLinearIcon from '@material-ui/icons/BlurLinear'
import { NavLink } from 'react-router-dom'
import React from 'react'
import TuneIcon from '@material-ui/icons/Tune'
import WebIcon from '@material-ui/icons/Web'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    appBar: {
      top: 'auto',
      bottom: 0,
    },
  })
)

const items = [
  { icon: <WebIcon />, href: 'overview' },
  {
    icon: <TuneIcon />,
    href: 'experiments',
  },
  {
    icon: <BlurLinearIcon />,
    href: 'events',
  },
  {
    icon: <ArchiveOutlinedIcon />,
    href: 'archives',
  },
]

const MobileNavigation = () => {
  const classes = useStyles()

  return (
    <AppBar className={classes.appBar} position="fixed" color="inherit">
      <Toolbar>
        {items.map((i) => (
          <NavLink key={i.href} to={i.href}>
            <IconButton color="primary">{i.icon}</IconButton>
          </NavLink>
        ))}
      </Toolbar>
    </AppBar>
  )
}

export default MobileNavigation
