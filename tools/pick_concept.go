// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"

	"tutor-mcp/engine"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PickConceptParams struct {
	Concept  string `json:"concept" jsonschema:"Le concept à épingler comme focus persistant. Vide pour clear le pin existant."`
	DomainID string `json:"domain_id,omitempty" jsonschema:"Domaine (défaut : dernier actif)"`
}

func registerPickConcept(server *mcp.Server, deps *Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "pick_concept",
		Description: "Épingle un concept comme focus prioritaire pour un domaine — override persistant du focus auto-calculé. Appeler quand l'apprenant clique un KC alternatif dans le cockpit ou demande explicitement de travailler un concept précis. Concept vide = clear le pin existant. Retourne l'OLMGraph mis à jour.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params PickConceptParams) (*mcp.CallToolResult, any, error) {
		learnerID, err := getLearnerID(ctx)
		if err != nil {
			deps.Logger.Error("pick_concept: auth failed", "err", err)
			r, _ := errorResult(err.Error())
			return r, nil, nil
		}

		domain, err := resolveDomain(deps.Store, learnerID, params.DomainID)
		if err != nil {
			deps.Logger.Error("pick_concept: resolve domain", "err", err, "learner", learnerID)
			r, _ := errorResult(err.Error())
			return r, nil, nil
		}

		// Validate concept exists in the domain (unless empty = clear).
		if params.Concept != "" {
			found := false
			for _, c := range domain.Graph.Concepts {
				if c == params.Concept {
					found = true
					break
				}
			}
			if !found {
				msg := fmt.Sprintf("concept %q inconnu dans le domaine %q", params.Concept, domain.Name)
				r, _ := errorResult(msg)
				return r, nil, nil
			}
		}

		if err := deps.Store.SetPinnedConcept(learnerID, domain.ID, params.Concept); err != nil {
			deps.Logger.Error("pick_concept: set pin", "err", err, "learner", learnerID)
			r, _ := errorResult(err.Error())
			return r, nil, nil
		}

		graph, err := engine.BuildOLMGraph(deps.Store, learnerID, domain.ID)
		if err != nil {
			deps.Logger.Error("pick_concept: build graph", "err", err, "learner", learnerID)
			r, _ := errorResult(err.Error())
			return r, nil, nil
		}

		r, _ := jsonResult(graph)
		return r, nil, nil
	})
}
