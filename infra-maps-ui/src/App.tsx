import { useInfraStore } from './store/infra'
import { useInfraTree } from './hooks/useInfraTree'
import { findNodeById } from './utils/tree'
import { GridView } from './views/GridView'
import { SidePanel } from './components/SidePanel'
import './App.css'

function App() {
  useInfraTree()

  const tree = useInfraStore((s) => s.tree)
  const isLoading = useInfraStore((s) => s.isLoading)
  const isError = useInfraStore((s) => s.isError)
  const errorMessage = useInfraStore((s) => s.errorMessage)
  const selectedNodeId = useInfraStore((s) => s.selectedNodeId)
  const selectNode = useInfraStore((s) => s.selectNode)
  const lastRefresh = useInfraStore((s) => s.lastRefresh)

  const selectedNode = selectedNodeId ? findNodeById(tree, selectedNodeId) : null

  return (
    <div className="app">
      <header className="toolbar">
        <span className="toolbar__logo">InfraMaps</span>
        <span className="toolbar__spacer" />
        {lastRefresh && (
          <span className="toolbar__refresh mono">MAJ {lastRefresh.toLocaleTimeString()}</span>
        )}
      </header>

      <main className="app__main">
        <GridView
          tree={tree}
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
