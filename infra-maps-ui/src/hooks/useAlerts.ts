import { useEffect } from 'react'
import { createAlertWebSocket } from '../api/websocket'
import { useInfraStore } from '../store/infra'

export function useAlerts() {
  useEffect(() => {
    const { addAlert, removeAlert } = useInfraStore.getState()
    const ws = createAlertWebSocket({
      onAlertFired: (alert) => {
        // Dédupliquer : le backend renvoie l'état complet à la reconnexion
        removeAlert(alert.id)
        addAlert(alert)
      },
      onAlertResolved: ({ id }) => removeAlert(id),
    })
    return () => ws.close()
  }, [])
}
