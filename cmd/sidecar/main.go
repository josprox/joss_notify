package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type request struct {
	Protocol string        `json:"protocol"`
	ID       string        `json:"id"`
	Method   string        `json:"method"`
	Args     []interface{} `json:"args"`
}

type response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

func main() {
	var req request
	if err := json.NewDecoder(io.LimitReader(os.Stdin, 8<<20)).Decode(&req); err != nil {
		write(response{Error: rpcError("BAD_REQUEST", err)})
		return
	}
	if req.Protocol != "joss-rpc-v1" || req.Method != "send" || len(req.Args) != 1 {
		write(response{ID: req.ID, Error: rpcError("BAD_REQUEST", fmt.Errorf("se requiere send(payload) sobre joss-rpc-v1"))})
		return
	}
	payload, ok := req.Args[0].(map[string]interface{})
	if !ok {
		write(response{ID: req.ID, Error: rpcError("BAD_PAYLOAD", fmt.Errorf("payload debe ser un objeto"))})
		return
	}
	result, err := dispatch(payload)
	if err != nil {
		write(response{ID: req.ID, Error: rpcError("NOTIFY_ERROR", err)})
		return
	}
	write(response{ID: req.ID, Result: result})
}

func dispatch(payload map[string]interface{}) (interface{}, error) {
	if endpoint := strings.TrimSpace(os.Getenv("NOTIFY_WEBHOOK_URL")); endpoint != "" {
		return sendWebhook(endpoint, payload)
	}
	if key := strings.TrimSpace(os.Getenv("FCM_SERVER_KEY")); key != "" {
		return sendFCMLegacy(key, payload)
	}
	return nil, fmt.Errorf("configure NOTIFY_WEBHOOK_URL o FCM_SERVER_KEY; las notificaciones in-app requieren un gateway que tenga acceso a la DB/WebSocket de la aplicación")
}

func sendWebhook(endpoint string, payload map[string]interface{}) (interface{}, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(os.Getenv("NOTIFY_WEBHOOK_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gateway HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var decoded interface{}
	if len(bytes.TrimSpace(data)) > 0 && json.Unmarshal(data, &decoded) == nil {
		return decoded, nil
	}
	return true, nil
}

func sendFCMLegacy(key string, payload map[string]interface{}) (interface{}, error) {
	target := ""
	if user := strings.TrimSpace(fmt.Sprint(payload["user"])); user != "" && user != "<nil>" {
		target = user
	} else if segment := strings.TrimSpace(fmt.Sprint(payload["segment"])); segment != "" && segment != "<nil>" {
		target = "/topics/" + strings.TrimPrefix(segment, "/topics/")
	}
	if target == "" {
		return nil, fmt.Errorf("FCM directo requiere user(token) o segment(topic)")
	}
	fcm := map[string]interface{}{
		"to":           target,
		"notification": map[string]interface{}{"title": payload["title"], "body": payload["message"]},
		"data":         payload,
	}
	body, _ := json.Marshal(fcm)
	req, _ := http.NewRequest(http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(body))
	req.Header.Set("Authorization", "key="+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("FCM HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return true, nil
	}
	return result, nil
}

func rpcError(code string, err error) map[string]string {
	return map[string]string{"code": code, "message": err.Error()}
}

func write(value response) { _ = json.NewEncoder(os.Stdout).Encode(value) }
