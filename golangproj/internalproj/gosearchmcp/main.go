// Package main implements an MCP server (std library only) that exposes
// a single tool "search_keyword" to query Baidu by keyword and return text for LLM analysis.
// Transports: stdio (newline-delimited JSON) or HTTP (POST body = JSON-RPC).
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// --- JSON-RPC 2.0 ---

type JsonRpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"` // absent for notifications
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JsonRpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JsonRpcErr `json:"error,omitempty"`
}

type JsonRpcErr struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// --- MCP result types ---

type InitializeResult struct {
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    ServerCaps `json:"capabilities"`
	ServerInfo      ServerInfo `json:"serverInfo"`
}

type ServerCaps struct {
	Tools *ToolsCap `json:"tools,omitempty"`
}

type ToolsCap struct {
	ListChanged bool `json:"listChanged"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ToolsListResult struct {
	Tools []ToolDef `json:"tools"`
}

type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Tool name and schema
const (
	toolName        = "search_keyword"
	toolDescription = "Search Baidu by keyword. Use when the user wants to look up a term, definition, or any topic. Returns page content for the model to analyze."
)

var searchKeywordSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"keyword": map[string]interface{}{
			"type":        "string",
			"description": "Search keyword or phrase (e.g. a word to look up)",
		},
	},
	"required": []string{"keyword"},
}

// dispatch handles one JSON-RPC request and returns the response (nil for notifications).
func dispatch(ctx context.Context, req *JsonRpcRequest) *JsonRpcResponse {
	// Notifications have no id; do not respond
	if req.ID == nil {
		if req.Method == "notifications/initialized" || req.Method == "initialized" {
			return nil
		}
		// Ignore other notifications
		return nil
	}

	switch req.Method {
	case "initialize":
		return handleInitialize(req)
	case "tools/list":
		return handleToolsList(req)
	case "tools/call":
		return handleToolsCall(ctx, req)
	default:
		return errorResponse(req.ID, -32601, "Method not found: "+req.Method)
	}
}

func handleInitialize(req *JsonRpcRequest) *JsonRpcResponse {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCaps{
			Tools: &ToolsCap{ListChanged: true},
		},
		ServerInfo: ServerInfo{
			Name:    "baidu-search-mcp",
			Version: "1.0.0",
		},
	}
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func handleToolsList(req *JsonRpcRequest) *JsonRpcResponse {
	result := ToolsListResult{
		Tools: []ToolDef{{
			Name:        toolName,
			Description: toolDescription,
			InputSchema: searchKeywordSchema,
		}},
	}
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func handleToolsCall(ctx context.Context, req *JsonRpcRequest) *JsonRpcResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return errorResponse(req.ID, -32602, "Invalid params: "+err.Error())
		}
	}
	if params.Name != toolName {
		return errorResponse(req.ID, -32602, "Unknown tool: "+params.Name)
	}
	keyword, _ := params.Arguments["keyword"].(string)
	if keyword == "" {
		return errorResponse(req.ID, -32602, "Missing required argument: keyword")
	}

	text, err := fetchBaiduSearch(ctx, keyword)
	if err != nil {
		return &JsonRpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: CallToolResult{
				Content: []ContentItem{{Type: "text", Text: "Error: " + err.Error()}},
				IsError: true,
			},
		}
	}
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: CallToolResult{
			Content: []ContentItem{{Type: "text", Text: text}},
		},
	}
}

func errorResponse(id interface{}, code int, msg string) *JsonRpcResponse {
	return &JsonRpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &JsonRpcErr{Code: code, Message: msg},
	}
}

// fetchBaiduSearch GETs Baidu search page for keyword, returns cleaned text (stdlib only).
func fetchBaiduSearch(ctx context.Context, keyword string) (string, error) {
	u := "https://www.baidu.com/s?wd=" + url.QueryEscape(keyword)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	html := string(body)
	return html, nil
}

// runStdio reads newline-delimited JSON-RPC from stdin, writes responses to stdout.
func runStdio(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req JsonRpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(JsonRpcResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &JsonRpcErr{Code: -32700, Message: "Parse error: " + err.Error()},
			})
			continue
		}
		resp := dispatch(ctx, &req)
		if resp != nil {
			if err := enc.Encode(resp); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

// runHTTP starts a server; POST body = single JSON-RPC request, response = single JSON-RPC response.
func runHTTP(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var req JsonRpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(JsonRpcResponse{
				JSONRPC: "2.0",
				Error:   &JsonRpcErr{Code: -32700, Message: "Parse error"},
			})
			return
		}
		resp := dispatch(r.Context(), &req)
		w.Header().Set("Content-Type", "application/json")
		if resp != nil {
			_ = json.NewEncoder(w).Encode(resp)
		} else {
			// Notification: return empty or minimal ack
			_, _ = w.Write([]byte("{}"))
		}
	})

	server := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	return server.ListenAndServe()
}

func main() {
	transport := flag.String("transport", "http", "stdio or http")
	port := flag.String("port", "8080", "port for http transport")
	flag.Parse()

	ctx := context.Background()

	switch *transport {
	case "stdio":
		if err := runStdio(ctx); err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "http":
		addr := ":" + *port
		if err := runHTTP(ctx, addr); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown transport: %s (use stdio or http)\n", *transport)
		os.Exit(1)
	}
}
