package github

import (
    "encoding/base64"
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "io"
    "net"
    "net/http"
    "os/exec"
    "strings"
    "time"
)

type Client struct {
    Token      string
    Repo       string // owner/repo
    HTTPClient *http.Client
}

func NewClient(token, repo string) *Client {
    return &Client{
        Token: token,
        Repo:  repo,
        HTTPClient: &http.Client{Timeout: 600 * time.Second},
    }
}

func (c *Client) do(ctx context.Context, method, url string, body interface{}) ([]byte, int, error) {
    var reqBody io.Reader
    if body != nil {
        data, err := json.Marshal(body)
        if err != nil {
            return nil, 0, err
        }
        if method == "POST" && strings.Contains(url, "dispatches") {
            log.Printf("DEBUG do(): Dispatching to %s with body: %s", url, string(data))
        }
        reqBody = bytes.NewReader(data)
    }

    req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
    if err != nil {
        return nil, 0, err
    }
    req.Header.Set("Accept", "application/vnd.github+json")
    req.Header.Set("Authorization", "Bearer "+c.Token)
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, 0, err
    }
    defer resp.Body.Close()

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, resp.StatusCode, err
    }
    return data, resp.StatusCode, nil
}

// DispatchWorkflow creates a workflow_dispatch event
func (c *Client) DispatchWorkflow(ctx context.Context, workflowID string, inputs map[string]interface{}) error {
    url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/%s/dispatches", c.Repo, workflowID)
    log.Printf("DEBUG DispatchWorkflow: URL=%s, workflowID=%s", url, workflowID)
    payload := map[string]interface{}{
        "ref":    "main",
        "inputs": inputs,
    }
    respBody, status, err := c.do(ctx, http.MethodPost, url, payload)
    if err != nil {
        return err
    }
    if status != http.StatusNoContent {
        log.Printf("DEBUG DispatchWorkflow ERROR: status %d, response: %s", status, string(respBody))
        return fmt.Errorf("dispatch failed: status %d, response: %s", status, string(respBody))
    }
    return nil
}

// GetRunByID returns a workflow run by ID
func (c *Client) GetRun(ctx context.Context, runID string) (map[string]interface{}, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%s", c.Repo, runID)
    data, status, err := c.do(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    if status != http.StatusOK {
        return nil, fmt.Errorf("get run failed: status %d", status)
    }
    var result map[string]interface{}
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }
    return result, nil
}

// ListArtifacts returns artifacts for a run
func (c *Client) ListArtifacts(ctx context.Context, runID string) ([]map[string]interface{}, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%s/artifacts", c.Repo, runID)
    data, status, err := c.do(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    if status != http.StatusOK {
        return nil, fmt.Errorf("list artifacts failed: status %d", status)
    }
    var result struct {
        Artifacts []map[string]interface{} `json:"artifacts"`
    }
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }
    return result.Artifacts, nil
}

// DownloadArtifact downloads an artifact ZIP by URL
func (c *Client) DownloadArtifact(ctx context.Context, url string) ([]byte, error) {
	// Primary path: pure Go HTTP. Works on Android (no curl).
	data, err := c.downloadArtifactHTTP(ctx, url)
	if err == nil {
		return data, nil
	}
	// Fallback: shell curl. Only useful on Termux/desktop for DNS quirks.
	log.Printf("DEBUG: Go HTTP download failed: %v, trying curl fallback", err)
	cmd := exec.Command("curl", "-s", "-L", "-H", "Authorization: Bearer "+c.Token, url)
	output, err2 := cmd.Output()
	if err2 == nil && len(output) > 0 {
		return output, nil
	}
	return nil, err
}

func (c *Client) downloadArtifactHTTP(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Accept", "application/vnd.github+json")
    
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 10 {
                return fmt.Errorf("too many redirects")
            }
            // Keep Authorization on all redirects (GitHub → Azure)
            if req.Header.Get("Authorization") == "" {
                req.Header.Set("Authorization", "Bearer "+c.Token)
            }
            return nil
        },
        Timeout: 600 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        10,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     60 * time.Second,
            DialContext: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
                Resolver: &net.Resolver{
                    PreferGo: true,
                    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                        d := net.Dialer{Timeout: 10 * time.Second}
                        return d.DialContext(ctx, "udp", "8.8.8.8:53")
                    },
                },
            }).DialContext,
        },
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("download artifact failed: status %d", resp.StatusCode)
    }
    return io.ReadAll(resp.Body)
}

// ListWorkflowRuns lists recent workflow runs for the workflow
func (c *Client) ListWorkflowRuns(ctx context.Context, workflowID string) ([]map[string]interface{}, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/%s/runs?per_page=10", c.Repo, workflowID)
    data, status, err := c.do(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    if status != http.StatusOK {
        return nil, fmt.Errorf("list runs failed: status %d", status)
    }
    var result struct {
        WorkflowRuns []map[string]interface{} `json:"workflow_runs"`
    }
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }
    return result.WorkflowRuns, nil
}

// CreateRepo creates a private repository under the authenticated user
func (c *Client) CreateRepo(ctx context.Context, name, description string) (map[string]interface{}, error) {
    url := "https://api.github.com/user/repos"
    payload := map[string]interface{}{
        "name":        name,
        "description": description,
        "private":     true,
        "auto_init":   false,
    }
    data, status, err := c.do(ctx, http.MethodPost, url, payload)
    if err != nil {
        return nil, err
    }
    if status != http.StatusCreated {
        return nil, fmt.Errorf("create repo failed: status %d: %s", status, string(data))
    }
    var repo map[string]interface{}
    if err := json.Unmarshal(data, &repo); err != nil {
        return nil, err
    }
    return repo, nil
}

// PushFile creates or updates a file in a repo
func (c *Client) PushFile(ctx context.Context, repo, path, content, message string) error {
    return c.pushFileWithRetry(ctx, repo, path, content, message, 3)
}

func (c *Client) pushFileWithRetry(ctx context.Context, repo, path, content, message string, retries int) error {
    path = strings.TrimPrefix(path, "/")
    url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, path)
    log.Printf("DEBUG: PushFile URL: %s", url)
    
    encoded := base64.StdEncoding.EncodeToString([]byte(content))
    
    payload := map[string]interface{}{
        "message": message,
        "content": encoded,
        "branch":  "main",
    }
    
    var body []byte
    var status int
    var err error

    for attempt := 1; attempt <= retries; attempt++ {
        body, status, err = c.do(ctx, http.MethodPut, url, payload)
        if err == nil && (status == http.StatusCreated || status == http.StatusOK) {
            return nil
        }
        if attempt < retries {
            backoff := time.Duration(attempt*2) * time.Second
            time.Sleep(backoff)
            log.Printf("PushFile retry %d after %v (status: %d)", attempt, backoff, status)
        }
    }

    if err != nil {
        return fmt.Errorf("push file network error: %w", err)
    }
    return fmt.Errorf("push file failed after %d retries: status %d, body: %s", retries, status, string(body))
}
