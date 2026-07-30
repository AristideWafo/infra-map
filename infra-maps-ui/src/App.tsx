import { useInfraStore } from './store/infra'
import { useInfraTree } from './hooks/useInfraTree'
import { useConnections } from './hooks/useConnections'
import { useAlerts } from './hooks/useAlerts'
import { useTags } from './hooks/useTags'
import { findNodeById, flattenTree } from './utils/tree'
import { GridView } from './views/GridView'
import { SidePanel } from './components/SidePanel'
import { FilterBar } from './components/Toolbar/FilterBar'
import './App.css'

function App() {
  useInfraTree()
  useConnections()
  useAlerts()
  const tags = useTags()

  const tree = useInfraStore((s) => s.tree)
  const connections = useInfraStore((s) => s.connections)
  const alerts = useInfraStore((s) => s.alerts)
  const isLoading = useInfraStore((s) => s.isLoading)
  const isError = useInfraStore((s) => s.isError)
  const errorMessage = useInfraStore((s) => s.errorMessage)
  const selectedNodeId = useInfraStore((s) => s.selectedNodeId)
  const selectNode = useInfraStore((s) => s.selectNode)
  const lastRefresh = useInfraStore((s) => s.lastRefresh)
  const filterNamespace = useInfraStore((s) => s.filterNamespace)
  const filterStatus = useInfraStore((s) => s.filterStatus)
  const filterTag = useInfraStore((s) => s.filterTag)
  const setFilterNamespace = useInfraStore((s) => s.setFilterNamespace)
  const setFilterStatus = useInfraStore((s) => s.setFilterStatus)
  const setFilterTag = useInfraStore((s) => s.setFilterTag)

  const selectedNode = selectedNodeId ? findNodeById(tree, selectedNodeId) : null
  const alertingIds = new Set(alerts.map((a) => a.nodeId))
  const namespaces = [
    ...new Set(flattenTree(tree).flatMap((n) => (n.namespace ? [n.namespace] : []))),
  ].sort()

  return (
    <div className="app">
      <header className="toolbar">
        <span className="toolbar__logo">InfraMaps</span>
        <FilterBar
          namespaces={namespaces}
          tags={tags}
          filterNamespace={filterNamespace}
          filterStatus={filterStatus}
          filterTag={filterTag}
          onNamespaceChange={setFilterNamespace}
          onStatusChange={setFilterStatus}
          onTagChange={setFilterTag}
        />
        <span className="toolbar__spacer" />
        {lastRefresh && (
          <span className="toolbar__refresh mono">MAJ {lastRefresh.toLocaleTimeString()}</span>
        )}
      </header>

      {isError && tree && (
        <div className="stale-banner" role="alert">
          ⚠ Rafraîchissement en échec — données périmées affichées ({errorMessage})
        </div>
      )}

      <main className="app__main">
        <GridView
          tree={tree}
          connections={connections}
          alertingIds={alertingIds}
          isLoading={isLoading}
          isError={isError}
          errorMessage={errorMessage}
          selectedNodeId={selectedNodeId}
          onSelect={(node) => selectNode(node.id)}
        />
        {selectedNode && <SidePanel node={selectedNode} onClose={() => selectNode(null)} />}
      </main>
    </div>
  )
}

export default App
