package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/cloud-gt/ai-sensors/command"
	"github.com/google/uuid"
)

type TestClient struct {
	srv *Server
}

func newTestClient(srv *Server) *TestClient {
	return &TestClient{srv: srv}
}

type Response struct {
	StatusCode int
	Body       []byte
}

func (r *Response) Decode(v any) error {
	return json.Unmarshal(r.Body, v)
}

func (tc *TestClient) Do(method, path string, body any) *Response {
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	tc.srv.router.ServeHTTP(w, req)

	return &Response{
		StatusCode: w.Code,
		Body:       w.Body.Bytes(),
	}
}

func (tc *TestClient) CreateCommand(name, cmd, workDir string) (*command.Command, *Response) {
	resp := tc.Do(http.MethodPost, "/commands", map[string]string{
		"name":     name,
		"command":  cmd,
		"work_dir": workDir,
	})

	if resp.StatusCode != http.StatusCreated {
		return nil, resp
	}

	var created command.Command
	_ = resp.Decode(&created)
	return &created, resp
}

func (tc *TestClient) GetCommand(id uuid.UUID) (*command.Command, *Response) {
	resp := tc.Do(http.MethodGet, "/commands/"+id.String(), nil)

	if resp.StatusCode != http.StatusOK {
		return nil, resp
	}

	var cmd command.Command
	_ = resp.Decode(&cmd)
	return &cmd, resp
}

func (tc *TestClient) ListCommands() ([]command.Command, *Response) {
	resp := tc.Do(http.MethodGet, "/commands", nil)

	if resp.StatusCode != http.StatusOK {
		return nil, resp
	}

	var result struct {
		Commands []command.Command `json:"commands"`
	}
	_ = resp.Decode(&result)
	return result.Commands, resp
}

func (tc *TestClient) DeleteCommand(id uuid.UUID) *Response {
	return tc.Do(http.MethodDelete, "/commands/"+id.String(), nil)
}

func (tc *TestClient) StartCommand(id uuid.UUID) (bool, *Response) {
	resp := tc.Do(http.MethodPost, "/commands/"+id.String()+"/start", nil)

	if resp.StatusCode != http.StatusOK {
		return false, resp
	}

	var result struct {
		Started bool `json:"started"`
	}
	_ = resp.Decode(&result)
	return result.Started, resp
}

func (tc *TestClient) StopCommand(id uuid.UUID) *Response {
	return tc.Do(http.MethodPost, "/commands/"+id.String()+"/stop", nil)
}

func (tc *TestClient) GetStatus(id uuid.UUID) (string, *Response) {
	resp := tc.Do(http.MethodGet, "/commands/"+id.String()+"/status", nil)

	if resp.StatusCode != http.StatusOK {
		return "", resp
	}

	var result struct {
		Status string `json:"status"`
	}
	_ = resp.Decode(&result)
	return result.Status, resp
}

func (tc *TestClient) GetOutput(id uuid.UUID) ([]string, *Response) {
	resp := tc.Do(http.MethodGet, "/commands/"+id.String()+"/output", nil)

	if resp.StatusCode != http.StatusOK {
		return nil, resp
	}

	var result struct {
		Lines []string `json:"lines"`
	}
	_ = resp.Decode(&result)
	return result.Lines, resp
}

func (tc *TestClient) GetOutputLastN(id uuid.UUID, n int) ([]string, *Response) {
	resp := tc.Do(http.MethodGet, "/commands/"+id.String()+"/output?lines="+strconv.Itoa(n), nil)

	if resp.StatusCode != http.StatusOK {
		return nil, resp
	}

	var result struct {
		Lines []string `json:"lines"`
	}
	_ = resp.Decode(&result)
	return result.Lines, resp
}
