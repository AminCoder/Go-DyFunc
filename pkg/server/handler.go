package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	Registry "github.com/AminCoder/Go-DyFunc/pkg/registry"
)

var i_username, i_password string

// response is the standard structure for the response to the client.
type response struct {
	ID    interface{} `json:"id,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// http_handler processes multiple function calls in a batch
func http_handler(w http.ResponseWriter, req *http.Request, registry *Registry.Function_Registry) {
	var batch_requests []struct {
		ID        interface{}   `json:"id"`
		Func_Name string        `json:"func"`
		Args      []interface{} `json:"args"`
	}

	decoder := json.NewDecoder(req.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&batch_requests); err != nil {
		http.Error(w, "invalid json input: "+err.Error(), http.StatusBadRequest)
		return
	}
	err := registry.Invoke_Middlewares(batch_requests, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	batch_results := make(map[string]map[string]interface{})
	result_chan := make(chan struct {
		ID   string
		Data map[string]interface{}
	}, len(batch_requests))
	for i, req := range batch_requests {
		var id string
		if req.ID == nil {
			id = strconv.Itoa(i)
		} else {
			id = fmt.Sprintf("%v", req.ID)
		}
		go func(req struct {
			ID        interface{}   `json:"id"`
			Func_Name string        `json:"func"`
			Args      []interface{} `json:"args"`
		}, id string) {
			results, err := registry.Call(context.TODO(), req.Func_Name, req.Args...)
			data := make(map[string]interface{})
			if err != nil {
				data["error"] = err.Error()
			} else {
				data["data"] = results
			}
			result_chan <- struct {
				ID   string
				Data map[string]interface{}
			}{ID: id, Data: data}
		}(req, id)
	}

	for i := 0; i < len(batch_requests); i++ {
		result := <-result_chan
		batch_results[result.ID] = result.Data
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(batch_results); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// Run_HTTP_Server starts a new HTTP server with the given address and port
func Run_HTTP_Server(server_address string, pattern string, registry *Registry.Function_Registry) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		if ok, err := registry.Check_Authentication(req); ok {
			http_handler(w, req, registry)
		} else {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
	})
	log.Printf("Starting server on %s\n", server_address)
	if err := http.ListenAndServe(server_address, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
