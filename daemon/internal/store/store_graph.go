package store

import "strconv"

// RawNode is a knowledge-graph entity before layout.
type RawNode struct {
	ID      string
	Label   string
	Kind    string
	Context string
}

// RawEdge links two entity ids.
type RawEdge struct {
	From string
	To   string
}

// GraphData returns entities and relations for layout by the domain layer.
func (s *Store) GraphData() ([]RawNode, []RawEdge, error) {
	nrows, err := s.db.Query(`
		SELECT e.id, e.rotulo, e.tipo, COALESCE(c.nome,'')
		FROM entidades e LEFT JOIN contextos c ON c.id = e.contexto_id ORDER BY e.id`)
	if err != nil {
		return nil, nil, err
	}
	defer nrows.Close()
	var nodes []RawNode
	for nrows.Next() {
		var id int64
		var n RawNode
		if err := nrows.Scan(&id, &n.Label, &n.Kind, &n.Context); err != nil {
			return nil, nil, err
		}
		n.ID = strconv.FormatInt(id, 10)
		nodes = append(nodes, n)
	}
	if err := nrows.Err(); err != nil {
		return nil, nil, err
	}

	erows, err := s.db.Query(`SELECT origem_id, destino_id FROM relacoes`)
	if err != nil {
		return nil, nil, err
	}
	defer erows.Close()
	var edges []RawEdge
	for erows.Next() {
		var a, b int64
		if err := erows.Scan(&a, &b); err != nil {
			return nil, nil, err
		}
		edges = append(edges, RawEdge{From: strconv.FormatInt(a, 10), To: strconv.FormatInt(b, 10)})
	}
	return nodes, edges, erows.Err()
}

// AddEntity inserts a knowledge-graph node and returns its id.
func (s *Store) AddEntity(tipo, rotulo, contextName string) (string, error) {
	ctxID, _ := s.contextMeta(contextName)
	res, err := s.db.Exec(
		`INSERT INTO entidades (tipo, rotulo, contexto_id) VALUES (?, ?, ?)`,
		tipo, rotulo, nullID(ctxID))
	if err != nil {
		return "", err
	}
	id, _ := res.LastInsertId()
	return strconv.FormatInt(id, 10), nil
}
