import { useInfraStore } from './store/infra'
import { useInfraTree } from './hooks/useInfraTree'
import { useConnections } from './hooks/useConnections'
import { useAlerts } from './hooks/useAlerts'
import { useTags } from './hooks/useTags'
import { useEffect } from 'react'
import { findNodeById, flattenTree, getAncestors } from './utils/tree'
import { GridView } from './views/GridView'
import { SidePanel } from './components/SidePanel'
import { FilterBar } from './components/Toolbar/FilterBar'
import { Breadcrumb } from './components/Toolbar/Breadcrumb'
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
  const breadcrumbPath = selectedNodeId ? getAncestors(tree, selectedNodeId) : []
  const alertingIds = new Set(alerts.map((a) => a.nodeId))

  // URL partageable : ?node=<id> suit la sélection
  useEffect(() => {
    const url = new URL(window.location.href)
    if (selectedNodeId) url.searchParams.set('node', selectedNodeId)
    else url.searchParams.delete('node')
    window.history.replaceState(null, '', url)
  }, [selectedNodeId])

  // Restaurer la sélection depuis l'URL au premier arbre chargé
  useEffect(() => {
    if (!tree) return
    const fromUrl = new URLSearchParams(window.location.search).get('node')
    if (fromUrl && !selectedNodeId && findNodeById(tree, fromUrl)) {
      selectNode(fromUrl)
    }
  }, [tree, selectedNodeId, selectNode])

  // Échap ferme le panel
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') selectNode(null)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [selectNode])
  const namespaces = [
    ...new Set(flattenTree(tree).flatMap((n) => (n.namespace ? [n.namespace] : []))),
  ].sort()

  return (
    <div className="app">
      <header className="toolbar">
        <span className="toolbar__logo">InfraMaps</span>
        <Breadcrumb path={breadcrumbPath} onNavigate={selectNode} />
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
