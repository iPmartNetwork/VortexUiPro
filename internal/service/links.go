package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"vortexuipro/internal/database"
)

// ─── Share Link Generators ───────────────────────────────────────────

// BuildVmessLink creates a vmess:// share link for a client.
func BuildVmessLink(client *database.Client, ib *database.Inbound, stream map[string]any, remark string) string {
	obj := map[string]any{
		"v":    "2",
		"ps":   remark,
		"add":  ib.Listen,
		"port": ib.Port,
		"id":   client.ID,
		"aid":  "0",
		"type": "none",
	}

	if stream != nil {
		if network, ok := stream["network"].(string); ok {
			obj["net"] = network
		}
		if security, ok := stream["security"].(string); ok {
			obj["tls"] = security
		}
		// WS settings
		if ws, ok := stream["wsSettings"].(map[string]any); ok {
			if path, ok := ws["path"].(string); ok {
				obj["path"] = path
			}
			if headers, ok := ws["headers"].(map[string]any); ok {
				if host, ok := headers["Host"].(string); ok {
					obj["host"] = host
				}
			}
		}
		// TLS settings
		if tls, ok := stream["tlsSettings"].(map[string]any); ok {
			if sni, ok := tls["serverName"].(string); ok {
				obj["sni"] = sni
			}
			if fp, ok := tls["fingerprint"].(string); ok {
				obj["fp"] = fp
			}
		}
	}

	jsonStr, _ := json.Marshal(obj)
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonStr)
}

// BuildVlessLink creates a vless:// share link.
func BuildVlessLink(client *database.Client, ib *database.Inbound, stream map[string]any, remark string) string {
	address := ib.Listen
	port := ib.Port

	link := fmt.Sprintf("vless://%s@%s", client.ID, joinHostPort(address, port))
	params := make(map[string]string)

	if stream != nil {
		params["type"] = getString(stream, "network", "tcp")
		applyStreamParams(stream, params)

		if security, ok := stream["security"].(string); ok {
			switch security {
			case "tls":
				if tls, ok := stream["tlsSettings"].(map[string]any); ok {
					params["security"] = "tls"
					params["sni"] = getString(tls, "serverName", "")
					params["fp"] = getString(tls, "fingerprint", "")
					if alpn, ok := tls["alpn"].([]any); ok {
						var alpnStrs []string
						for _, a := range alpn {
							alpnStrs = append(alpnStrs, fmt.Sprint(a))
						}
						params["alpn"] = strings.Join(alpnStrs, ",")
					}
				}
			case "reality":
				if reality, ok := stream["realitySettings"].(map[string]any); ok {
					params["security"] = "reality"
					params["sni"] = getString(reality, "serverName", "")
					params["pbk"] = getString(reality, "publicKey", "")
					params["fp"] = getString(reality, "fingerprint", "")
					params["sid"] = getString(reality, "shortId", "")
				}
			default:
				params["security"] = "none"
			}
		} else {
			params["security"] = "none"
		}

		if client.Flow != "" {
			params["flow"] = client.Flow
		}
	}

	return buildLinkWithParams(link, params, remark)
}

// BuildTrojanLink creates a trojan:// share link.
func BuildTrojanLink(client *database.Client, ib *database.Inbound, stream map[string]any, remark string) string {
	password := url.QueryEscape(client.Password)
	link := fmt.Sprintf("trojan://%s@%s", password, joinHostPort(ib.Listen, ib.Port))
	params := make(map[string]string)

	if stream != nil {
		params["type"] = getString(stream, "network", "tcp")
		if ws, ok := stream["wsSettings"].(map[string]any); ok {
			params["path"] = getString(ws, "path", "")
			if headers, ok := ws["headers"].(map[string]any); ok {
				params["host"] = getString(headers, "Host", "")
			}
		}
		applyTLSParams(stream, params)
	}

	return buildLinkWithParams(link, params, remark)
}

// BuildShadowsocksLink creates an ss:// share link.
func BuildShadowsocksLink(client *database.Client, ib *database.Inbound, settings map[string]any, stream map[string]any, remark string) string {
	method := getString(settings, "method", "chacha20-poly1305")
	password := getString(settings, "password", "")

	// UserInfo = base64(method:password)
	userInfo := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", method, password)))

	link := fmt.Sprintf("ss://%s@%s", userInfo, joinHostPort(ib.Listen, ib.Port))
	params := make(map[string]string)

	if stream != nil {
		params["type"] = getString(stream, "network", "tcp")
		if ws, ok := stream["wsSettings"].(map[string]any); ok {
			params["path"] = getString(ws, "path", "")
			if headers, ok := ws["headers"].(map[string]any); ok {
				params["host"] = getString(headers, "Host", "")
			}
		}
	}

	return buildLinkWithParams(link, params, remark)
}

// BuildHysteria2Link creates a hysteria2:// share link.
func BuildHysteria2Link(client *database.Client, ib *database.Inbound, settings map[string]any, stream map[string]any, remark string) string {
	auth := url.QueryEscape(client.Password)
	protocol := "hysteria2"
	if v, ok := settings["version"].(float64); ok && v == 1 {
		protocol = "hysteria"
	}

	link := fmt.Sprintf("%s://%s@%s", protocol, auth, joinHostPort(ib.Listen, ib.Port))
	params := make(map[string]string)

	params["security"] = "tls"
	if stream != nil {
		if tls, ok := stream["tlsSettings"].(map[string]any); ok {
			params["sni"] = getString(tls, "serverName", "")
			params["fp"] = getString(tls, "fingerprint", "")
			if alpn, ok := tls["alpn"].([]any); ok {
				var alpnStrs []string
				for _, a := range alpn {
					alpnStrs = append(alpnStrs, fmt.Sprint(a))
				}
				params["alpn"] = strings.Join(alpnStrs, ",")
			}
		}
	}

	return buildLinkWithParams(link, params, remark)
}

// BuildWireGuardLink creates a wireguard:// share link.
func BuildWireGuardLink(client *database.Client, ib *database.Inbound, settings map[string]any, remark string) string {
	if client.PrivateKey == "" {
		return ""
	}

	link := fmt.Sprintf("wireguard://%s@%s", client.PrivateKey, joinHostPort(ib.Listen, ib.Port))
	params := make(map[string]string)

	if pubKey, ok := settings["publicKey"].(string); ok && pubKey != "" {
		params["publickey"] = pubKey
	}
	if client.AllowedIPs != "" {
		params["address"] = client.AllowedIPs
	}
	if mtu, ok := settings["mtu"].(float64); ok && mtu > 0 {
		params["mtu"] = strconv.Itoa(int(mtu))
	}
	if dns, ok := settings["dns"].(string); ok && dns != "" {
		params["dns"] = dns
	}
	if client.KeepAlive > 0 {
		params["keepalive"] = strconv.Itoa(client.KeepAlive)
	}

	return buildLinkWithParams(link, params, remark)
}

// BuildMTProtoLink creates a tg:// proxy link for MTProto.
func BuildMTProtoLink(client *database.Client, ib *database.Inbound, remark string) string {
	params := map[string]string{
		"server": ib.Listen,
		"port":   fmt.Sprintf("%d", ib.Port),
		"secret": client.Secret,
	}
	return buildLinkWithParams("tg://proxy", params, "")
}

// ─── Helpers ─────────────────────────────────────────────────────────

func joinHostPort(host string, port int) string {
	host = strings.Trim(host, "[]")
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func buildLinkWithParams(link string, params map[string]string, remark string) string {
	if len(params) == 0 && remark == "" {
		return link
	}

	var queryParts []string
	for k, v := range params {
		if v != "" {
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
		}
	}

	if len(queryParts) > 0 {
		link += "?" + strings.Join(queryParts, "&")
	}
	if remark != "" {
		link += "#" + url.QueryEscape(remark)
	}
	return link
}

func applyStreamParams(stream map[string]any, params map[string]string) {
	network := getString(stream, "network", "tcp")
	switch network {
	case "ws":
		if ws, ok := stream["wsSettings"].(map[string]any); ok {
			params["path"] = getString(ws, "path", "")
			if headers, ok := ws["headers"].(map[string]any); ok {
				params["host"] = getString(headers, "Host", "")
			}
		}
	case "grpc":
		if grpc, ok := stream["grpcSettings"].(map[string]any); ok {
			params["serviceName"] = getString(grpc, "serviceName", "")
			params["authority"] = getString(grpc, "authority", "")
		}
	case "tcp":
		if tcp, ok := stream["tcpSettings"].(map[string]any); ok {
			if header, ok := tcp["header"].(map[string]any); ok {
				if typeStr, ok := header["type"].(string); ok && typeStr == "http" {
					params["headerType"] = "http"
				}
			}
		}
	}
}

func applyTLSParams(stream map[string]any, params map[string]string) {
	if security, ok := stream["security"].(string); ok {
		switch security {
		case "tls":
			params["security"] = "tls"
			if tls, ok := stream["tlsSettings"].(map[string]any); ok {
				params["sni"] = getString(tls, "serverName", "")
				params["fp"] = getString(tls, "fingerprint", "")
			}
		case "reality":
			params["security"] = "reality"
			if reality, ok := stream["realitySettings"].(map[string]any); ok {
				params["sni"] = getString(reality, "serverName", "")
				params["pbk"] = getString(reality, "publicKey", "")
				params["fp"] = getString(reality, "fingerprint", "")
			}
		}
	}
}

func getString(m map[string]any, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

// ApplyRemarkTemplate applies remark template variables.
func ApplyRemarkTemplate(template string, client *database.Client, ib *database.Inbound, user *database.User) string {
	if template == "" {
		template = "{inbound_remark} - {client_email}"
	}

	replacer := strings.NewReplacer(
		"{client_email}", client.Email,
		"{client_id}", client.ID,
		"{inbound_remark}", ib.Remark,
		"{inbound_tag}", ib.Tag,
		"{inbound_protocol}", string(ib.Protocol),
		"{inbound_port}", strconv.Itoa(ib.Port),
		"{inbound_host}", ib.Listen,
	)
	result := replacer.Replace(template)

	if user != nil {
		result = strings.NewReplacer(
			"{user_data_limit}", formatBytes(user.DataLimit),
			"{user_expiry}", fmt.Sprintf("%d", user.ExpiryTime),
			"{user_status}", user.Status,
		).Replace(result)
	}

	// Random generators
	if strings.Contains(result, "{random_email}") {
		result = strings.ReplaceAll(result, "{random_email}", fmt.Sprintf("%x@x.com", sha256.Sum256([]byte(client.ID)))[:20])
	}
	if strings.Contains(result, "{random_uuid}") {
		hash := sha256.Sum256([]byte(client.ID))
		uuid := hex.EncodeToString(hash[:16])
		uuid = uuid[:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:32]
		result = strings.ReplaceAll(result, "{random_uuid}", uuid)
	}

	return result
}

func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	size := float64(bytes)
	for size >= 1024 && i < len(units)-1 {
		size /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", size, units[i])
}
