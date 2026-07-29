package xray

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// ─── gRPC Service Constants ───────────────────────────────────────────

const (
	statsServiceName   = "xray.app.stats.command.StatsService"
	handlerServiceName = "xray.app.proxyman.command.HandlerService"
	routerServiceName  = "xray.app.router.command.RoutingService"
)

// ─── Method Paths ─────────────────────────────────────────────────────

var (
	methodStatsQueryStats = fmt.Sprintf("/%s/QueryStats", statsServiceName)
	methodStatsGetStats   = fmt.Sprintf("/%s/GetStats", statsServiceName)
	methodStatsGetUsers   = fmt.Sprintf("/%s/GetUsersStats", statsServiceName)

	methodHandlerAddInbound    = fmt.Sprintf("/%s/AddInbound", handlerServiceName)
	methodHandlerRemInbound    = fmt.Sprintf("/%s/RemoveInbound", handlerServiceName)
	methodHandlerAddOutbound   = fmt.Sprintf("/%s/AddOutbound", handlerServiceName)
	methodHandlerRemOutbound   = fmt.Sprintf("/%s/RemoveOutbound", handlerServiceName)
	methodHandlerAlterInbound  = fmt.Sprintf("/%s/AlterInbound", handlerServiceName)

	methodRouterAddRule        = fmt.Sprintf("/%s/AddRule", routerServiceName)
	methodRouterGetBalancer    = fmt.Sprintf("/%s/GetBalancerInfo", routerServiceName)
	methodRouterOverride       = fmt.Sprintf("/%s/OverrideBalancerTarget", routerServiceName)
	methodRouterTestRoute      = fmt.Sprintf("/%s/TestRoute", routerServiceName)
)

// ─── Raw Codec ────────────────────────────────────────────────────────

type rawCodec struct{}

func (rawCodec) Marshal(v any) ([]byte, error) {
	switch v := v.(type) {
	case []byte:
		return v, nil
	case *[]byte:
		return *v, nil
	case proto.Message:
		return proto.Marshal(v)
	default:
		return nil, fmt.Errorf("rawCodec: unsupported type %T", v)
	}
}

func (rawCodec) Unmarshal(data []byte, v any) error {
	switch v := v.(type) {
	case *[]byte:
		*v = data
		return nil
	case proto.Message:
		return proto.Unmarshal(data, v)
	default:
		return fmt.Errorf("rawCodec: unsupported type %T", v)
	}
}

func (rawCodec) Name() string { return "raw" }

// ─── gRPC Client ───────────────────────────────────────────────────────

type grpcClient struct {
	conn   *grpc.ClientConn
	addr   string
	dialed bool
}

func newGRPCClient(addr string) *grpcClient {
	return &grpcClient{addr: addr}
}

func (c *grpcClient) connect(ctx context.Context) error {
	if c.dialed && c.conn != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(rawCodec{}),
			grpc.MaxCallRecvMsgSize(1024*1024),
		),
	)
	if err != nil {
		return fmt.Errorf("xray gRPC dial %s: %w", c.addr, err)
	}
	c.conn = conn
	c.dialed = true
	return nil
}

func (c *grpcClient) close() {
	if c.conn != nil {
		c.conn.Close()
		c.dialed = false
	}
}

// ─── Protobuf Wire Encoding Helpers ────────────────────────────────────

// encodeString appends a string field (wire type 2 / LEN).
func encodeString(tag protowire.Number, s string) []byte {
	var b []byte
	b = protowire.AppendTag(b, tag, protowire.BytesType)
	b = protowire.AppendString(b, s)
	return b
}

// encodeVarint appends a varint field (wire type 0).
func encodeVarint(tag protowire.Number, v int64) []byte {
	var b []byte
	b = protowire.AppendTag(b, tag, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(v))
	return b
}

// encodeBytes appends a bytes field (wire type 2 / LEN).
func encodeBytes(tag protowire.Number, data []byte) []byte {
	var b []byte
	b = protowire.AppendTag(b, tag, protowire.BytesType)
	b = protowire.AppendBytes(b, data)
	return b
}

// encodeBool appends a bool field (wire type 0 / varint).
func encodeBool(tag protowire.Number, v bool) []byte {
	var b []byte
	b = protowire.AppendTag(b, tag, protowire.VarintType)
	b = protowire.AppendVarint(b, protowire.EncodeBool(v))
	return b
}

// encodeInt32 appends an int32 field (wire type 0 / varint).
func encodeInt32(tag protowire.Number, v int32) []byte {
	var b []byte
	b = protowire.AppendTag(b, tag, protowire.VarintType)
	b = protowire.AppendVarint(b, uint64(v))
	return b
}

// ─── StatsService Protobuf Messages ────────────────────────────────────

// QueryStatsRequest: pattern=1 (string), reset=2 (bool)
func encodeQueryStatsRequest(pattern string, reset bool) []byte {
	return append(encodeString(1, pattern), encodeBool(2, reset)...)
}

// GetStatsRequest: name=1 (string), reset=2 (bool)
func encodeGetStatsRequest(name string, reset bool) []byte {
	return append(encodeString(1, name), encodeBool(2, reset)...)
}

// GetUsersStatsRequest: (empty message, no fields)
func encodeGetUsersStatsRequest() []byte { return nil }

// Parse QueryStatsResponse: stats=1 (repeated Stat, bytes)
// Stat: name=1 (string), value=2 (int64)
func parseQueryStatsResponse(data []byte) (map[string]int64, error) {
	result := make(map[string]int64)
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return result, fmt.Errorf("parse tag: %v", protowire.ParseError(n))
		}
		data = data[n:]

		if num == 1 && wtype == protowire.BytesType {
			val, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return result, fmt.Errorf("parse stat bytes: %v", protowire.ParseError(n))
			}
			data = data[n:]
			name, value := parseStat(val)
			if name != "" {
				result[name] = value
			}
		} else {
			n = protowire.ConsumeFieldValue(num, wtype, data)
			if n < 0 {
				return result, fmt.Errorf("skip field: %v", protowire.ParseError(n))
			}
			data = data[n:]
		}
	}
	return result, nil
}

func parseStat(data []byte) (name string, value int64) {
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return "", 0
		}
		data = data[n:]

		switch num {
		case 1: // name (string)
			if wtype == protowire.BytesType {
				val, n := protowire.ConsumeBytes(data)
				if n > 0 {
					name = string(val)
					data = data[n:]
				}
			} else {
				n = protowire.ConsumeFieldValue(num, wtype, data)
				if n > 0 {
					data = data[n:]
				}
			}
		case 2: // value (int64)
			if wtype == protowire.VarintType {
				val, n := protowire.ConsumeVarint(data)
				if n > 0 {
					value = int64(val)
					data = data[n:]
				}
			} else {
				n = protowire.ConsumeFieldValue(num, wtype, data)
				if n > 0 {
					data = data[n:]
				}
			}
		default:
			n = protowire.ConsumeFieldValue(num, wtype, data)
			if n > 0 {
				data = data[n:]
			}
		}
	}
	return name, value
}

// ─── HandlerService Protobuf Messages ─────────────────────────────────

// AddInboundRequest: inbound=1 (bytes, serialized InboundConfig)
func encodeAddInboundRequest(inboundJSON []byte) []byte {
	return encodeString(1, string(inboundJSON))
}

// RemoveInboundRequest: tag=1 (string)
func encodeRemoveInboundRequest(tag string) []byte {
	return encodeString(1, tag)
}

// AddOutboundRequest: outbound=1 (bytes, serialized OutboundConfig)
func encodeAddOutboundRequest(outboundJSON []byte) []byte {
	return encodeString(1, string(outboundJSON))
}

// RemoveOutboundRequest: tag=1 (string)
func encodeRemoveOutboundRequest(tag string) []byte {
	return encodeString(1, tag)
}

// AlterInboundRequest: tag=1 (string), operation=2 (bytes, Any)
// AddUserOperation: user=1 (User proto)
// User: level=1 (int32), email=2 (string), account=3 (Any, bytes)
// RemoveUserOperation: email=1 (string)
func encodeAddUserOperation(tag, email, protocol, userID string) []byte {
	// Build the protocol-specific account Any message
	var accBytes []byte
	switch protocol {
	case "vmess":
		// VMessAccount: id=1 (string), security=2 (string), level=3 (int32)
		accBytes = append(accBytes, encodeString(1, userID)...)
		accBytes = append(accBytes, encodeString(2, "auto")...)
		accBytes = append(accBytes, encodeInt32(3, 0)...)
	case "vless":
		// VLESSAccount: id=1 (string), flow=2 (string), encryption=3 (string), level=4 (int32)
		accBytes = append(accBytes, encodeString(1, userID)...)
		accBytes = append(accBytes, encodeString(2, "")...)
		accBytes = append(accBytes, encodeString(3, "none")...)
		accBytes = append(accBytes, encodeInt32(4, 0)...)
	case "trojan":
		// TrojanAccount: password=1 (string), flow=2 (string), level=3 (int32)
		accBytes = append(accBytes, encodeString(1, userID)...)
		accBytes = append(accBytes, encodeInt32(3, 0)...)
	default:
		// Default to VMess account
		accBytes = append(accBytes, encodeString(1, userID)...)
		accBytes = append(accBytes, encodeString(2, "auto")...)
		accBytes = append(accBytes, encodeInt32(3, 0)...)
	}

	// Build google.protobuf.Any wrapper
	typeURL := "type.googleapis.com/xray.proxy." + protocol + ".Account"
	anyBytes := append(encodeString(1, typeURL), encodeBytes(2, accBytes)...)

	// Build User proto
	userBytes := encodeInt32(1, 0) // level=0
	userBytes = append(userBytes, encodeString(2, email)...)
	userBytes = append(userBytes, encodeBytes(3, anyBytes)...)

	// Build AddUserOperation
	opBytes := encodeBytes(1, userBytes)

	// Wrap operation in Any type
	opTypeURL := "type.googleapis.com/xray.app.proxyman.command.AddUserOperation"
	anyOp := append(encodeString(1, opTypeURL), encodeBytes(2, opBytes)...)

	// Build AlterInboundRequest
	return append(encodeString(1, tag), encodeBytes(2, anyOp)...)
}

func encodeRemoveUserOperation(tag, email string) []byte {
	// Build RemoveUserOperation
	opBytes := encodeString(1, email)

	// Wrap in Any
	opTypeURL := "type.googleapis.com/xray.app.proxyman.command.RemoveUserOperation"
	anyOp := append(encodeString(1, opTypeURL), encodeBytes(2, opBytes)...)

	// Build AlterInboundRequest
	return append(encodeString(1, tag), encodeBytes(2, anyOp)...)
}

// ─── RoutingService Protobuf Messages ─────────────────────────────────

// AddRuleRequest: config=1 (bytes), should_append=2 (bool)
func encodeAddRuleRequest(routingJSON []byte) []byte {
	return append(encodeBytes(1, routingJSON), encodeBool(2, false)...)
}

// GetBalancerInfoRequest: tag=1 (string)
func encodeGetBalancerInfoRequest(tag string) []byte {
	return encodeString(1, tag)
}

// OverrideBalancerTargetRequest: balancer_tag=1 (string), target=2 (string)
func encodeOverrideBalancerRequest(balancerTag, target string) []byte {
	return append(encodeString(1, balancerTag), encodeString(2, target)...)
}

// TestRouteRequest: routing_context=1 (RoutingContext), publish_result=2 (bool)
// RoutingContext: inbound_tag=1 (string), network=2 (int32), target_domain=3 (string),
//   target_port=4 (uint32), target_ips=5 (repeated bytes), protocol=6 (string), user=7 (string)
func encodeTestRouteRequest(req *RouteTestRequest) []byte {
	// Build RoutingContext
	ctxBytes := encodeString(1, req.InboundTag)

	// Network: tcp=0, udp=1
	network := int32(0)
	if strings.EqualFold(req.Network, "udp") {
		network = 1
	}
	ctxBytes = append(ctxBytes, encodeInt32(2, network)...)
	ctxBytes = append(ctxBytes, encodeString(3, req.Domain)...)
	ctxBytes = append(ctxBytes, encodeVarint(4, int64(req.Port))...)
	ctxBytes = append(ctxBytes, encodeString(6, req.Protocol)...)
	ctxBytes = append(ctxBytes, encodeString(7, req.Email)...)

	// Build TestRouteRequest
	return append(encodeBytes(1, ctxBytes), encodeBool(2, false)...)
}

// parseTestRouteResponse parses TestRouteResponse
// TestRouteResponse: outbound_tag=1 (string), outbound_group_tags=2 (repeated string)
func parseTestRouteResponse(data []byte) *RouteTestResult {
	result := &RouteTestResult{}
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return result
		}
		data = data[n:]

		switch num {
		case 1: // outbound_tag
			if wtype == protowire.BytesType {
				val, n := protowire.ConsumeBytes(data)
				if n > 0 {
					result.OutboundTag = string(val)
					result.Matched = true
					data = data[n:]
					continue
				}
			}
		case 2: // outbound_group_tags (repeated string)
			if wtype == protowire.BytesType {
				val, n := protowire.ConsumeBytes(data)
				if n > 0 {
					result.GroupTags = append(result.GroupTags, string(val))
					data = data[n:]
					continue
				}
			}
		}
		n = protowire.ConsumeFieldValue(num, wtype, data)
		if n > 0 {
			data = data[n:]
		}
	}
	return result
}

// parseGetBalancerResponse parses GetBalancerInfoResponse
// GetBalancerInfoResponse: balancer=1 (Balancer, bytes)
// Balancer: override=1 (Override, bytes), principle_target=2 (PrincipleTarget, bytes)
// Override: target=1 (string)
// PrincipleTarget: tag=1 (repeated string)
func parseGetBalancerResponse(data []byte) *BalancerInfo {
	info := &BalancerInfo{}
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return info
		}
		data = data[n:]

		if num == 1 && wtype == protowire.BytesType {
			balBytes, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return info
			}
			data = data[n:]
			info = parseBalancer(balBytes, info.Tag)
		} else {
			n = protowire.ConsumeFieldValue(num, wtype, data)
			if n > 0 {
				data = data[n:]
			}
		}
	}
	return info
}

func parseBalancer(data []byte, tag string) *BalancerInfo {
	info := &BalancerInfo{Tag: tag}
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return info
		}
		data = data[n:]

		switch num {
		case 1: // override
			if wtype == protowire.BytesType {
				ovBytes, n := protowire.ConsumeBytes(data)
				if n > 0 {
					info.Override = parseOverride(ovBytes)
					data = data[n:]
					continue
				}
			}
		case 2: // principle_target
			if wtype == protowire.BytesType {
				ptBytes, n := protowire.ConsumeBytes(data)
				if n > 0 {
					info.Selected = parsePrincipleTarget(ptBytes)
					data = data[n:]
					continue
				}
			}
		}
		n = protowire.ConsumeFieldValue(num, wtype, data)
		if n > 0 {
			data = data[n:]
		}
	}
	return info
}

func parseOverride(data []byte) string {
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return ""
		}
		data = data[n:]
		if num == 1 && wtype == protowire.BytesType {
			val, n := protowire.ConsumeBytes(data)
			if n > 0 {
				return string(val)
			}
		}
		n = protowire.ConsumeFieldValue(num, wtype, data)
		if n > 0 {
			data = data[n:]
		}
	}
	return ""
}

func parsePrincipleTarget(data []byte) []string {
	var tags []string
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return tags
		}
		data = data[n:]
		if num == 1 && wtype == protowire.BytesType {
			val, n := protowire.ConsumeBytes(data)
			if n > 0 {
				tags = append(tags, string(val))
				data = data[n:]
				continue
			}
		}
		n = protowire.ConsumeFieldValue(num, wtype, data)
		if n > 0 {
			data = data[n:]
		}
	}
	return tags
}

// ─── Stat Name Parsing ────────────────────────────────────────────────

// extractTagFromStat extracts the tag from an xray stat name like "inbound>>>[tag]>>>traffic>>>uplink"
func extractTagFromStat(name string) string {
	parts := strings.Split(name, ">>>")
	if len(parts) >= 3 {
		return parts[1]
	}
	return ""
}

// extractEmailFromStat extracts the email from a user stat name like "user>>>[email]>>>traffic>>>uplink"
func extractEmailFromStat(name string) string {
	parts := strings.Split(name, ">>>")
	if len(parts) >= 3 && parts[0] == "user" {
		return parts[1]
	}
	return ""
}

// statToTrafficStats converts a map of stat names to grouped Traffic entries.
func statToTrafficStats(stats map[string]int64) map[string]*Traffic {
	result := make(map[string]*Traffic)
	for name, value := range stats {
		tag := extractTagFromStat(name)
		if tag == "" || tag == "api" {
			continue
		}
		if _, ok := result[tag]; !ok {
			result[tag] = &Traffic{Tag: tag}
		}
		if strings.HasSuffix(name, "uplink") {
			result[tag].Up = value
		} else if strings.HasSuffix(name, "downlink") {
			result[tag].Down = value
		}
	}
	return result
}

// getClientTrafficFromStats extracts per-client traffic from user>>> stats.
func getClientTrafficFromStats(stats map[string]int64) []ClientTraffic {
	emailMap := make(map[string]*ClientTraffic)
	for name, value := range stats {
		if !strings.HasPrefix(name, "user>>>") {
			continue
		}
		parts := strings.Split(name, ">>>")
		if len(parts) >= 4 && parts[3] == "traffic" {
			email := parts[1]
			if _, ok := emailMap[email]; !ok {
				emailMap[email] = &ClientTraffic{Email: email}
			}
			if strings.HasSuffix(name, "uplink") {
				emailMap[email].Up = value
			} else if strings.HasSuffix(name, "downlink") {
				emailMap[email].Down = value
			}
		}
	}
	result := make([]ClientTraffic, 0, len(emailMap))
	for _, ct := range emailMap {
		result = append(result, *ct)
	}
	return result
}

// ─── High-Level gRPC Methods ─────────────────────────────────────────

func (c *grpcClient) invoke(ctx context.Context, method string, req []byte) ([]byte, error) {
	if err := c.connect(ctx); err != nil {
		return nil, err
	}
	var resp []byte
	if err := c.conn.Invoke(ctx, method, &req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ─── StatsService ─────────────────────────────────────────────────────

func (c *grpcClient) queryStats(ctx context.Context, pattern string, reset bool) (map[string]int64, error) {
	req := encodeQueryStatsRequest(pattern, reset)
	resp, err := c.invoke(ctx, methodStatsQueryStats, req)
	if err != nil {
		return nil, fmt.Errorf("QueryStats: %w", err)
	}
	return parseQueryStatsResponse(resp)
}

func (c *grpcClient) getStats(ctx context.Context, name string, reset bool) (int64, error) {
	req := encodeGetStatsRequest(name, reset)
	resp, err := c.invoke(ctx, methodStatsGetStats, req)
	if err != nil {
		return 0, fmt.Errorf("GetStats(%s): %w", name, err)
	}
	_, value := parseStat(resp)
	return value, nil
}

func (c *grpcClient) getUsersStats(ctx context.Context) ([]OnlineUser, error) {
	req := encodeGetUsersStatsRequest()
	resp, err := c.invoke(ctx, methodStatsGetUsers, req)
	if err != nil {
		return nil, fmt.Errorf("GetUsersStats: %w", err)
	}
	return parseUsersStatsResponse(resp)
}

func parseUsersStatsResponse(data []byte) ([]OnlineUser, error) {
	// GetUsersStatsResponse: users=1 (repeated User, bytes)
	var users []OnlineUser
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return users, nil
		}
		data = data[n:]

		if num == 1 && wtype == protowire.BytesType {
			userBytes, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return users, nil
			}
			data = data[n:]
			user := parseOnlineUser(userBytes)
			if user.Email != "" {
				users = append(users, user)
			}
		} else {
			n = protowire.ConsumeFieldValue(num, wtype, data)
			if n > 0 {
				data = data[n:]
			}
		}
	}
	return users, nil
}

func parseOnlineUser(data []byte) OnlineUser {
	var user OnlineUser
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return user
		}
		data = data[n:]

		switch num {
		case 1: // email (string)
			if wtype == protowire.BytesType {
				val, n := protowire.ConsumeBytes(data)
				if n > 0 {
					user.Email = string(val)
					data = data[n:]
					continue
				}
			}
		case 2: // ips (repeated IP, bytes)
			if wtype == protowire.BytesType {
				ipBytes, n := protowire.ConsumeBytes(data)
				if n > 0 {
					ip := parseOnlineIP(ipBytes)
					if ip.IP != "" {
						user.IPs = append(user.IPs, ip)
					}
					data = data[n:]
					continue
				}
			}
		}
		n = protowire.ConsumeFieldValue(num, wtype, data)
		if n > 0 {
			data = data[n:]
		}
	}
	return user
}

func parseOnlineIP(data []byte) OnlineIP {
	var ip OnlineIP
	for len(data) > 0 {
		num, wtype, n := protowire.ConsumeTag(data)
		if n < 0 {
			return ip
		}
		data = data[n:]

		switch num {
		case 1: // ip (string)
			if wtype == protowire.BytesType {
				val, n := protowire.ConsumeBytes(data)
				if n > 0 {
					ip.IP = string(val)
					data = data[n:]
					continue
				}
			}
		case 2: // last_seen (int64)
			if wtype == protowire.VarintType {
				val, n := protowire.ConsumeVarint(data)
				if n > 0 {
					ip.LastSeen = int64(val)
					data = data[n:]
					continue
				}
			}
		}
		n = protowire.ConsumeFieldValue(num, wtype, data)
		if n > 0 {
			data = data[n:]
		}
	}
	return ip
}

// ─── HandlerService ───────────────────────────────────────────────────

func (c *grpcClient) addInbound(ctx context.Context, inboundJSON []byte) error {
	req := encodeAddInboundRequest(inboundJSON)
	_, err := c.invoke(ctx, methodHandlerAddInbound, req)
	if err != nil {
		return fmt.Errorf("AddInbound: %w", err)
	}
	return nil
}

func (c *grpcClient) removeInbound(ctx context.Context, tag string) error {
	req := encodeRemoveInboundRequest(tag)
	_, err := c.invoke(ctx, methodHandlerRemInbound, req)
	if err != nil {
		return fmt.Errorf("RemoveInbound(%s): %w", tag, err)
	}
	return nil
}

func (c *grpcClient) addOutbound(ctx context.Context, outboundJSON []byte) error {
	req := encodeAddOutboundRequest(outboundJSON)
	_, err := c.invoke(ctx, methodHandlerAddOutbound, req)
	if err != nil {
		return fmt.Errorf("AddOutbound: %w", err)
	}
	return nil
}

func (c *grpcClient) removeOutbound(ctx context.Context, tag string) error {
	req := encodeRemoveOutboundRequest(tag)
	_, err := c.invoke(ctx, methodHandlerRemOutbound, req)
	if err != nil {
		return fmt.Errorf("RemoveOutbound(%s): %w", tag, err)
	}
	return nil
}

func (c *grpcClient) addUser(ctx context.Context, tag, email, protocol, userID string) error {
	req := encodeAddUserOperation(tag, email, protocol, userID)
	_, err := c.invoke(ctx, methodHandlerAlterInbound, req)
	if err != nil {
		return fmt.Errorf("AddUser(%s, %s): %w", tag, email, err)
	}
	return nil
}

func (c *grpcClient) removeUser(ctx context.Context, tag, email string) error {
	req := encodeRemoveUserOperation(tag, email)
	_, err := c.invoke(ctx, methodHandlerAlterInbound, req)
	if err != nil {
		return fmt.Errorf("RemoveUser(%s, %s): %w", tag, email, err)
	}
	return nil
}

// ─── RoutingService ──────────────────────────────────────────────────

func (c *grpcClient) applyRouting(ctx context.Context, routingJSON []byte) error {
	req := encodeAddRuleRequest(routingJSON)
	_, err := c.invoke(ctx, methodRouterAddRule, req)
	if err != nil {
		return fmt.Errorf("ApplyRouting: %w", err)
	}
	return nil
}

func (c *grpcClient) getBalancerInfo(ctx context.Context, tag string) (*BalancerInfo, error) {
	req := encodeGetBalancerInfoRequest(tag)
	resp, err := c.invoke(ctx, methodRouterGetBalancer, req)
	if err != nil {
		return nil, fmt.Errorf("GetBalancerInfo(%s): %w", tag, err)
	}
	info := parseGetBalancerResponse(resp)
	info.Tag = tag
	return info, nil
}

func (c *grpcClient) setBalancerTarget(ctx context.Context, balancerTag, target string) error {
	req := encodeOverrideBalancerRequest(balancerTag, target)
	_, err := c.invoke(ctx, methodRouterOverride, req)
	if err != nil {
		return fmt.Errorf("OverrideBalancerTarget(%s, %s): %w", balancerTag, target, err)
	}
	return nil
}

func (c *grpcClient) testRoute(ctx context.Context, req *RouteTestRequest) (*RouteTestResult, error) {
	reqBytes := encodeTestRouteRequest(req)
	resp, err := c.invoke(ctx, methodRouterTestRoute, reqBytes)
	if err != nil {
		// "not enough information" means no rule matched
		if strings.Contains(strings.ToLower(err.Error()), "not enough information") {
			return &RouteTestResult{Matched: false}, nil
		}
		return nil, fmt.Errorf("TestRoute: %w", err)
	}
	return parseTestRouteResponse(resp), nil
}
