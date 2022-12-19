package pipeline

import "log"

func PrintCollection(col *Collection) {
	for _, v := range col.Graph.Nodes {
		log.Println("node:", v.Value.Name)
		for _, v := range v.Value.Graph.Nodes {
			log.Println("  node:", v.Value.Name)
		}
		for _, e := range v.Value.Graph.Edges {
			for _, v := range e {
				log.Println("  edge:", v.From.Value.Name, "->", v.To.Value.Name)
			}
		}
	}
	for _, e := range col.Graph.Edges {
		for _, v := range e {
			log.Println("edge:", v.From.Value.Name, "->", v.To.Value.Name)
		}
	}
}
