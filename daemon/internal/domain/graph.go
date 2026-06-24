package domain

import (
	"math"

	"aqnod/internal/model"
	"aqnod/internal/store"
)

// BuildGraph lays out the knowledge graph: a central companion node, the
// contexts in a ring around it, and each context's entities clustered nearby.
// Positions are normalised to [0,1] for the canvas renderer.
func BuildGraph(s *store.Store, companion string) (model.Graph, error) {
	rawNodes, rawEdges, err := s.GraphData()
	if err != nil {
		return model.Graph{}, err
	}
	ctxs, _ := s.Contexts()
	colorOf := map[string]string{}
	for _, c := range ctxs {
		colorOf[c.Label] = c.Color
	}

	// Partition nodes.
	var contextNodes []store.RawNode
	childrenOf := map[string][]store.RawNode{}
	byID := map[string]store.RawNode{}
	for _, n := range rawNodes {
		byID[n.ID] = n
		if n.Kind == "context" {
			contextNodes = append(contextNodes, n)
		}
	}
	for _, e := range rawEdges {
		// seed relation is child -> parent ('pertence_a')
		if parent, ok := byID[e.To]; ok && parent.Kind == "context" {
			childrenOf[e.To] = append(childrenOf[e.To], byID[e.From])
		}
	}

	if companion == "" {
		companion = "Aqno"
	}
	center := model.GraphNode{ID: "__persona__", Label: companion, X: 0.5, Y: 0.5, Color: "violet", Kind: "context", Size: 17}
	out := model.Graph{Nodes: []model.GraphNode{center}}

	nCtx := len(contextNodes)
	for i, c := range contextNodes {
		ang := 2 * math.Pi * float64(i) / math.Max(1, float64(nCtx))
		cx := 0.5 + 0.3*math.Cos(ang)
		cy := 0.5 + 0.3*math.Sin(ang)
		col := colorOf[c.Label]
		if col == "" {
			col = "violet"
		}
		out.Nodes = append(out.Nodes, model.GraphNode{
			ID: c.ID, Label: c.Label, X: clamp01(cx), Y: clamp01(cy), Color: col, Kind: "context", Size: 13,
		})
		out.Edges = append(out.Edges, model.GraphEdge{From: center.ID, To: c.ID})

		kids := childrenOf[c.ID]
		for j, k := range kids {
			ka := ang + (float64(j)-float64(len(kids)-1)/2)*0.5
			kx := cx + 0.13*math.Cos(ka)
			ky := cy + 0.13*math.Sin(ka)
			out.Nodes = append(out.Nodes, model.GraphNode{
				ID: k.ID, Label: k.Label, X: clamp01(kx), Y: clamp01(ky), Color: col, Kind: k.Kind,
			})
			out.Edges = append(out.Edges, model.GraphEdge{From: c.ID, To: k.ID})
		}
	}
	return out, nil
}

func clamp01(v float64) float64 {
	if v < 0.04 {
		return 0.04
	}
	if v > 0.96 {
		return 0.96
	}
	return v
}
