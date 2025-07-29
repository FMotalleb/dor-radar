import React, { useEffect, useRef, useState } from 'react';
import { RefreshCw, Loader, AlertCircle, Info } from 'lucide-react';

interface Node {
  id: string;
  name: string;
  x?: number;
  y?: number;
  fx?: number | null;
  fy?: number | null;
}

interface Connection {
  source: string | Node;
  target: string | Node;
  strength?: number;
}

interface ApiResponse {
  nodes: Array<{ id: number; name: string }>;
  connections: Array<{ source: number; target: number; strength?: number }>;
}

const NetworkGraph: React.FC = () => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [connections, setConnections] = useState<Connection[]>([]);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  // Mock API function - replace with your actual API endpoint
  const fetchNetworkData = async (): Promise<ApiResponse> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Mock data - replace this with actual fetch call
    return {
      nodes: [
        { id: 0, name: "local" },
        { id: 1, name: "database" },
        { id: 2, name: "api-server" },
        { id: 3, name: "load-balancer" },
        { id: 4, name: "cache-redis" },
        { id: 5, name: "auth-service" },
        { id: 6, name: "file-storage" },
        { id: 7, name: "monitoring" }
      ],
      connections: [
        { source: 0, target: 1, strength: 0.9 },
        { source: 0, target: 2, strength: 0.8 },
        { source: 2, target: 3, strength: 0.7 },
        { source: 3, target: 4, strength: 0.6 },
        { source: 2, target: 5, strength: 0.8 },
        { source: 0, target: 6, strength: 0.5 },
        { source: 2, target: 7, strength: 0.4 },
        { source: 4, target: 7, strength: 0.3 }
      ]
    };
    
    // Uncomment and modify this for actual API call:
    // const response = await fetch('YOUR_API_ENDPOINT');
    // if (!response.ok) {
    //   throw new Error(`HTTP error! status: ${response.status}`);
    // }
    // return await response.json();
  };

  const loadNetworkData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const data = await fetchNetworkData();
      
      // Transform API data to internal format
      const transformedNodes: Node[] = data.nodes.map(node => ({
        id: node.id.toString(),
        name: node.name
      }));
      
      const transformedConnections: Connection[] = data.connections.map(conn => ({
        source: conn.source.toString(),
        target: conn.target.toString(),
        strength: conn.strength || 0.5
      }));
      
      setNodes(transformedNodes);
      setConnections(transformedConnections);
      setLastUpdated(new Date());
      setSelectedNode(null);
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load network data');
    } finally {
      setLoading(false);
    }
  };

  // Load data on component mount
  useEffect(() => {
    loadNetworkData();
  }, []);

  useEffect(() => {
    if (!svgRef.current || nodes.length === 0) return;

    // Clear previous content
    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    const width = 800;
    const height = 600;

    // Create main group for zooming
    const g = svg.append('g');

    // Add zoom behavior
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.3, 3])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    svg.call(zoom);

    // Create force simulation
    const simulation = d3.forceSimulation<Node>(nodes)
      .force('link', d3.forceLink<Node, Connection>(connections)
        .id((d) => d.id)
        .distance(120)
        .strength((d: Connection) => d.strength || 0.5))
      .force('charge', d3.forceManyBody().strength(-400))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(40));

    // Create gradient definitions
    const defs = svg.append('defs');
    
    // Node gradient
    const nodeGradient = defs.append('radialGradient')
      .attr('id', 'nodeGradient')
      .attr('cx', '30%')
      .attr('cy', '30%');
    
    nodeGradient.append('stop')
      .attr('offset', '0%')
      .attr('stop-color', '#60A5FA')
      .attr('stop-opacity', 1);
    
    nodeGradient.append('stop')
      .attr('offset', '100%')
      .attr('stop-color', '#1E40AF')
      .attr('stop-opacity', 1);

    // Selected node gradient
    const selectedGradient = defs.append('radialGradient')
      .attr('id', 'selectedGradient')
      .attr('cx', '30%')
      .attr('cy', '30%');
    
    selectedGradient.append('stop')
      .attr('offset', '0%')
      .attr('stop-color', '#FCD34D')
      .attr('stop-opacity', 1);
    
    selectedGradient.append('stop')
      .attr('offset', '100%')
      .attr('stop-color', '#F59E0B')
      .attr('stop-opacity', 1);

    // Create arrow marker
    defs.append('marker')
      .attr('id', 'arrowhead')
      .attr('viewBox', '-0 -5 10 10')
      .attr('refX', 30)
      .attr('refY', 0)
      .attr('orient', 'auto')
      .attr('markerWidth', 8)
      .attr('markerHeight', 8)
      .attr('xoverflow', 'visible')
      .append('path')
      .attr('d', 'M 0,-5 L 10 ,0 L 0,5')
      .attr('fill', '#64748B')
      .attr('stroke', 'none');

    // Create connections
    const link = g.append('g')
      .attr('class', 'links')
      .selectAll('line')
      .data(connections)
      .enter().append('line')
      .attr('stroke', '#64748B')
      .attr('stroke-opacity', 0.6)
      .attr('stroke-width', (d: Connection) => Math.sqrt((d.strength || 0.5) * 8))
      .attr('marker-end', 'url(#arrowhead)');

    // Create nodes
    const node = g.append('g')
      .attr('class', 'nodes')
      .selectAll('g')
      .data(nodes)
      .enter().append('g')
      .attr('class', 'node')
      .style('cursor', 'pointer')
      .call(d3.drag<SVGGElement, Node>()
        .on('start', (event, d) => {
          if (!event.active) simulation.alphaTarget(0.3).restart();
          d.fx = d.x;
          d.fy = d.y;
        })
        .on('drag', (event, d) => {
          d.fx = event.x;
          d.fy = event.y;
        })
        .on('end', (event, d) => {
          if (!event.active) simulation.alphaTarget(0);
          d.fx = null;
          d.fy = null;
        }));

    // Add circles to nodes
    node.append('circle')
      .attr('r', 25)
      .attr('fill', (d: Node) => selectedNode === d.id ? 'url(#selectedGradient)' : 'url(#nodeGradient)')
      .attr('stroke', '#1E293B')
      .attr('stroke-width', 2)
      .style('filter', 'drop-shadow(0 4px 8px rgba(0,0,0,0.2))')
      .on('mouseover', function(event, d) {
        d3.select(this)
          .transition()
          .duration(200)
          .attr('r', 30)
          .style('filter', 'drop-shadow(0 6px 12px rgba(0,0,0,0.3))');
      })
      .on('mouseout', function(event, d) {
        d3.select(this)
          .transition()
          .duration(200)
          .attr('r', 25)
          .style('filter', 'drop-shadow(0 4px 8px rgba(0,0,0,0.2))');
      })
      .on('click', (event, d) => {
        setSelectedNode(selectedNode === d.id ? null : d.id);
      });

    // Add labels to nodes
    node.append('text')
      .text((d: Node) => d.name)
      .attr('dy', 40)
      .attr('text-anchor', 'middle')
      .attr('font-family', 'Inter, system-ui, sans-serif')
      .attr('font-size', '12px')
      .attr('font-weight', '600')
      .attr('fill', '#1E293B')
      .style('pointer-events', 'none')
      .style('user-select', 'none');

    // Update positions on simulation tick
    simulation.on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y);

      node
        .attr('transform', (d: Node) => `translate(${d.x},${d.y})`);
    });

    // Update node colors when selection changes
    node.select('circle')
      .attr('fill', (d: Node) => selectedNode === d.id ? 'url(#selectedGradient)' : 'url(#nodeGradient)');

  }, [nodes, connections, selectedNode]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900 flex items-center justify-center">
        <div className="text-center">
          <Loader className="w-12 h-12 text-blue-400 animate-spin mx-auto mb-4" />
          <h2 className="text-2xl font-semibold text-white mb-2">Loading Network Data</h2>
          <p className="text-blue-200">Fetching nodes and connections...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900 flex items-center justify-center">
        <div className="text-center max-w-md mx-auto p-6">
          <AlertCircle className="w-12 h-12 text-red-400 mx-auto mb-4" />
          <h2 className="text-2xl font-semibold text-white mb-2">Failed to Load Network</h2>
          <p className="text-red-200 mb-6">{error}</p>
          <button
            onClick={loadNetworkData}
            className="bg-blue-600 hover:bg-blue-700 text-white py-3 px-6 rounded-lg transition-all duration-200 flex items-center justify-center gap-2 font-medium shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 mx-auto"
          >
            <RefreshCw className="w-5 h-5" />
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-slate-900 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-white mb-2">
            Network Topology Visualization
          </h1>
          <p className="text-blue-200 text-lg">
            Real-time network graph powered by API data
          </p>
        </div>

        {/* Main Content */}
        <div className="flex flex-col lg:flex-row gap-6">
          {/* Graph Container */}
          <div className="flex-1">
            <div className="bg-white/10 backdrop-blur-sm rounded-2xl border border-white/20 p-6 shadow-2xl">
              <svg
                ref={svgRef}
                width="100%"
                height="600"
                viewBox="0 0 800 600"
                className="w-full h-auto border rounded-lg bg-gradient-to-br from-slate-50 to-slate-100"
              />
            </div>
          </div>

          {/* Info Panel */}
          <div className="lg:w-80">
            <div className="bg-white/10 backdrop-blur-sm rounded-2xl border border-white/20 p-6 shadow-2xl">
              <h2 className="text-xl font-semibold text-white mb-6 flex items-center gap-2">
                <Info className="w-5 h-5" />
                Network Information
              </h2>

              {/* Statistics */}
              <div className="grid grid-cols-2 gap-4 mb-6">
                <div className="bg-blue-500/20 rounded-lg p-3 border border-blue-400/30">
                  <div className="text-blue-200 text-sm">Nodes</div>
                  <div className="text-white text-2xl font-bold">{nodes.length}</div>
                </div>
                <div className="bg-green-500/20 rounded-lg p-3 border border-green-400/30">
                  <div className="text-green-200 text-sm">Connections</div>
                  <div className="text-white text-2xl font-bold">{connections.length}</div>
                </div>
              </div>

              {/* Refresh Button */}
              <button
                onClick={loadNetworkData}
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white py-3 px-4 rounded-lg transition-all duration-200 flex items-center justify-center gap-2 font-medium shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 disabled:transform-none mb-6"
              >
                <RefreshCw className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} />
                Refresh Data
              </button>

              {/* Last Updated */}
              {lastUpdated && (
                <div className="mb-6 p-3 bg-white/5 rounded-lg border border-white/10">
                  <div className="text-white text-sm font-medium">Last Updated</div>
                  <div className="text-blue-200 text-xs">
                    {lastUpdated.toLocaleString()}
                  </div>
                </div>
              )}

              {/* Selected Node Info */}
              {selectedNode && (
                <div className="mb-6 p-4 bg-yellow-500/20 rounded-lg border border-yellow-400/30">
                  <h3 className="text-yellow-200 text-sm font-medium mb-2">Selected Node</h3>
                  <div className="text-white font-semibold">
                    {nodes.find(n => n.id === selectedNode)?.name}
                  </div>
                  <div className="text-yellow-200 text-sm mt-1">
                    ID: {selectedNode}
                  </div>
                </div>
              )}

              {/* Instructions */}
              <div className="p-4 bg-white/5 rounded-lg border border-white/10">
                <h3 className="text-white text-sm font-medium mb-2">Instructions</h3>
                <ul className="text-blue-200 text-sm space-y-1">
                  <li>• Drag nodes to reposition them</li>
                  <li>• Click nodes to select/deselect</li>
                  <li>• Use mouse wheel to zoom</li>
                  <li>• Drag background to pan</li>
                  <li>• Click refresh to reload from API</li>
                </ul>
              </div>

              {/* API Info */}
              <div className="mt-4 p-4 bg-purple-500/20 rounded-lg border border-purple-400/30">
                <h3 className="text-purple-200 text-sm font-medium mb-2">API Integration</h3>
                <div className="text-purple-100 text-xs">
                  Data is fetched from API endpoint and automatically transformed for visualization.
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

function App() {
  return <NetworkGraph />;
}

export default App;