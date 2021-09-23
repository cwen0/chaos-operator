import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogProps,
  DialogTitle,
} from '@material-ui/core'

import T from 'components/T'

interface ConfirmDialogProps {
  open: boolean
  close?: () => void
  title: string | JSX.Element
  description?: string
  onConfirm?: () => void
  dialogProps?: Omit<DialogProps, 'open'>
}

const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  close,
  title,
  description,
  onConfirm,
  children,
  dialogProps,
}) => {
  const handleConfirm = () => {
    typeof onConfirm === 'function' && onConfirm()
    typeof close === 'function' && close()
  }

  return (
    <Dialog open={open} onClose={close} PaperProps={{ sx: { minWidth: 300 } }} {...dialogProps}>
      <DialogTitle sx={{ p: 3 }}>{title}</DialogTitle>
      <DialogContent sx={{ p: 3 }}>
        {children ? children : description ? <DialogContentText>{description}</DialogContentText> : null}
      </DialogContent>

      {onConfirm && (
        <DialogActions>
          <Button size="small" onClick={close}>
            {T('common.cancel')}
          </Button>
          <Button variant="contained" color="primary" size="small" autoFocus disableFocusRipple onClick={handleConfirm}>
            {T('common.confirm')}
          </Button>
        </DialogActions>
      )}
    </Dialog>
  )
}

export default ConfirmDialog
