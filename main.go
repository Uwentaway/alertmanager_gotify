package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

type Alert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

type AlertManagerPayload struct {
	Alerts []Alert `json:"alerts"`
}

type GotifyPayload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// 从环境变量获取 Gotify 的 URL 和 Token
var gotifyURL = os.Getenv("GOTIFY_URL")
var gotifyToken = os.Getenv("GOTIFY_TOKEN")

// alertTemplate 定义了新的警报模板
const alertTemplate = `
{{ range . }} 
{{ if eq .Status "firing" }}**[⚠️告警]**
告警名称: {{ .Labels.alertname }}
开始时间: {{ .StartsAt | formatTime }}
结束时间: {{ .StartsAt | formatTime }}
实例: {{ .Labels.instance }}
IP: {{ .Annotations.ip }}
描述: {{ .Annotations.description }}
{{ else }}**[✅恢复]**
告警名称: {{ .Labels.alertname }}
开始时间: {{ .StartsAt | formatTime }}
结束时间: {{ .EndsAt | formatTime }}
实例: {{ .Labels.instance }}
IP: {{ .Annotations.ip }}
描述: {{ .Annotations.description }}
{{ end }}
{{ end }}
`

// formatTime 是一个自定义的模板函数，用于格式化时间为亚洲上海时区
func formatTime(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Println("Error loading Shanghai timezone:", err)
		return t.Format(time.RFC3339) // 如果加载时区失败，使用默认格式
	}
	return t.In(loc).Format("2006-01-02 15:04:05") // 格式化为 "YYYY-MM-DD HH:mm:ss"
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// 解码 Alertmanager 发送的 JSON 数据
	var payload AlertManagerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to decode JSON payload", http.StatusBadRequest)
		log.Println("Error decoding JSON:", err)
		return
	}

	// 注册自定义的模板函数
	funcMap := template.FuncMap{
		"formatTime": formatTime,
	}

	// 渲染警报消息
	tmpl, err := template.New("alert").Funcs(funcMap).Parse(alertTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		log.Println("Error parsing template:", err)
		return
	}

	var message bytes.Buffer
	if err := tmpl.Execute(&message, payload.Alerts); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Println("Error rendering template:", err)
		return
	}

	// 创建 Gotify 负载
	gotifyPayload := GotifyPayload{
		Title:    "Prometheus Alert",
		Message:  message.String(),
		Priority: 5,
	}

	// 发送警报到 Gotify
	if err := sendToGotify(gotifyPayload); err != nil {
		http.Error(w, "Failed to send alert to Gotify", http.StatusInternalServerError)
		log.Println("Error sending to Gotify:", err)
		return
	}

	log.Println("Alert sent to Gotify successfully")
	w.WriteHeader(http.StatusOK)
}

// sendToGotify 将警报消息发送到 Gotify
func sendToGotify(payload GotifyPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, gotifyURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", gotifyToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Gotify returned non-200 status: %s", resp.Status)
	}

	return nil
}

func main() {
	// 检查环境变量是否配置
	if gotifyURL == "" || gotifyToken == "" {
		log.Fatal("GOTIFY_URL and GOTIFY_TOKEN must be set in environment variables")
	}

	http.HandleFunc("/webhook", webhookHandler)

	port := ":9110"
	log.Printf("Starting server on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}

