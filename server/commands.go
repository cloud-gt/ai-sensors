package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/cloud-gt/ai-sensors/command"
	"github.com/cloud-gt/ai-sensors/manager"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CommandsAPI struct {
	store   *command.Store
	manager *manager.Manager
}

func NewCommandsAPI(store *command.Store, mgr *manager.Manager) *CommandsAPI {
	return &CommandsAPI{
		store:   store,
		manager: mgr,
	}
}

func (api *CommandsAPI) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", api.handleList)
	r.Post("/", api.handleCreate)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", api.handleGet)
		r.Delete("/", api.handleDelete)
		r.Post("/start", api.handleStart)
		r.Post("/stop", api.handleStop)
		r.Get("/status", api.handleStatus)
		r.Get("/output", api.handleOutput)
	})
	return r
}

func (api *CommandsAPI) handleList(w http.ResponseWriter, r *http.Request) {
	commands, err := api.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"commands": commands})
}

func (api *CommandsAPI) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Command string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Command == "" {
		writeError(w, http.StatusBadRequest, "command is required")
		return
	}

	existing, err := api.store.GetByName(req.Name)
	if err == nil && existing.ID != uuid.Nil {
		writeError(w, http.StatusConflict, "command already exists")
		return
	}

	cmd, err := api.store.Create(command.Command{
		Name:    req.Name,
		Command: req.Command,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, cmd)
}

func (api *CommandsAPI) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid command ID")
		return
	}

	cmd, err := api.store.Get(id)
	if err != nil {
		if errors.Is(err, command.ErrNotFound) {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, cmd)
}

func (api *CommandsAPI) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid command ID")
		return
	}

	status, err := api.manager.Status(id)
	if err == nil && status == manager.StatusRunning {
		writeError(w, http.StatusConflict, "cannot delete running command")
		return
	}

	err = api.store.Delete(id)
	if err != nil {
		if errors.Is(err, command.ErrNotFound) {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *CommandsAPI) handleStart(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid command ID")
		return
	}

	started, err := api.manager.Start(r.Context(), id)
	if err != nil {
		if errors.Is(err, manager.ErrCommandNotFound) {
			writeError(w, http.StatusNotFound, "command not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"started": started})
}

func (api *CommandsAPI) handleStop(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid command ID")
		return
	}

	err = api.manager.Stop(id)
	if err != nil {
		if errors.Is(err, manager.ErrNotRunning) {
			writeError(w, http.StatusNotFound, "command not running")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (api *CommandsAPI) handleStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid command ID")
		return
	}

	status, err := api.manager.Status(id)
	if err != nil {
		if errors.Is(err, manager.ErrNotRunning) {
			writeError(w, http.StatusNotFound, "command not running")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": string(status)})
}

func (api *CommandsAPI) handleOutput(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid command ID")
		return
	}

	linesParam := r.URL.Query().Get("lines")
	var lines []string

	if linesParam != "" {
		n, err := strconv.Atoi(linesParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "lines must be a positive integer")
			return
		}
		if n < 0 {
			writeError(w, http.StatusBadRequest, "lines must be a positive integer")
			return
		}
		lines, err = api.manager.OutputLastN(id, n)
		if err != nil {
			if errors.Is(err, manager.ErrNotRunning) {
				writeError(w, http.StatusNotFound, "command not running")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	} else {
		lines, err = api.manager.Output(id)
		if err != nil {
			if errors.Is(err, manager.ErrNotRunning) {
				writeError(w, http.StatusNotFound, "command not running")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string][]string{"lines": lines})
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
