package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/thejff/go-chucklechain/key"
	"github.com/thejff/go-chucklechain/structs"
)

type API interface {
	Start() error
}

type api struct {
	server *http.Server
	key    key.Key[any]
	cfg    structs.DsConfig
}

func New(cfg structs.DsConfig, k key.Key[any]) API {
	a := &api{
		cfg: cfg,
		key: k,
	}

	h := http.NewServeMux()

	h.Handle("/self", a.self())

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", cfg.Self.ApiPort),
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	a.server = s

	return a
}

func (a *api) Start() error {
	fmt.Printf("API listening on %s\n", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *api) self() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := structs.SelfResp{
			Name:      a.cfg.Self.Name,
			PublicKey: a.key.GetPublicPEM(),
			Neighbours: structs.SelfNeighbour{
				Count: 0,
				Nodes: []interface{}{},
			},
		}

		bResp, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)

			w.WriteHeader(http.StatusInternalServerError)

			if _, err := w.Write([]byte(`{"error": "failed to marshal response"}`)); err != nil {
				log.Println(err)
			}
			return
		}

		if _, err := w.Write(bResp); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error": "failed to write response"}`)); err != nil {
				log.Println(err)
			}
			return
		}
	}
}
