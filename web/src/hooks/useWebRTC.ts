import { useEffect, useRef, useState, useCallback } from 'react'

// ─── Types ───────────────────────────────────────────────────────────

export interface RTCMessage {
  type: 'offer' | 'answer' | 'ice_candidate' | 'join' | 'leave' | 'heartbeat'
  from_id: string
  to_id?: string
  session_id?: string
  sdp?: string
  candidate?: string
  timestamp: number
}

export interface RTCPeerData {
  id: string
  connection: RTCPeerConnection
  channel: RTCDataChannel | null
  status: 'connecting' | 'connected' | 'disconnected'
  latency?: number
}

interface UseRTCConfig {
  peerId?: string
  iceServers?: RTCIceServer[]
  onMessage?: (peerId: string, data: string) => void
  onPeerConnected?: (peerId: string) => void
  onPeerDisconnected?: (peerId: string) => void
}

// ─── WebRTC Signaling Hook ──────────────────────────────────────────

export function useWebRTC(config: UseRTCConfig = {}) {
  const [peers, setPeers] = useState<Map<string, RTCPeerData>>(new Map())
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const peersRef = useRef<Map<string, RTCPeerData>>(new Map())

  const defaultIceServers: RTCIceServer[] = [
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' },
  ]

  const iceServers = config.iceServers || defaultIceServers

  // ─── WebSocket Signaling ─────────────────────────────────────────
  const connectSignaling = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/webrtc/signal/ws`

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      setConnected(true)
      setError(null)
      // Join the signaling room
      sendMessage({ type: 'join', from_id: config.peerId || 'client', timestamp: Date.now() })
    }

    ws.onclose = () => {
      setConnected(false)
      // Auto-reconnect after 3 seconds
      setTimeout(() => connectSignaling(), 3000)
    }

    ws.onerror = () => {
      setError('WebSocket connection error')
      setConnected(false)
    }

    ws.onmessage = (event) => {
      try {
        const msg: RTCMessage = JSON.parse(event.data)
        handleSignalingMessage(msg)
      } catch (e) {
        console.error('[WebRTC] Invalid message:', e)
      }
    }
  }, [config.peerId])

  // ─── Handle Signaling Messages ────────────────────────────────────
  const handleSignalingMessage = async (msg: RTCMessage) => {
    switch (msg.type) {
      case 'offer':
        await handleOffer(msg)
        break
      case 'answer':
        await handleAnswer(msg)
        break
      case 'ice_candidate':
        await handleICECandidate(msg)
        break
      case 'heartbeat':
        // Ping-pong
        break
    }
  }

  // ─── Create Peer Connection ───────────────────────────────────────
  const createPeerConnection = (peerId: string): RTCPeerConnection => {
    const pc = new RTCPeerConnection({ iceServers })
    const dataChannel = pc.createDataChannel('vortexuipro-mesh', {
      ordered: true,
    })

    dataChannel.onopen = () => {
      updatePeerState(peerId, 'connected')
      config.onPeerConnected?.(peerId)
    }

    dataChannel.onclose = () => {
      updatePeerState(peerId, 'disconnected')
      config.onPeerDisconnected?.(peerId)
    }

    dataChannel.onmessage = (event) => {
      config.onMessage?.(peerId, event.data)
    }

    pc.onicecandidate = (event) => {
      if (event.candidate) {
        sendMessage({
          type: 'ice_candidate',
          from_id: config.peerId || 'client',
          to_id: peerId,
          candidate: JSON.stringify(event.candidate),
          timestamp: Date.now(),
        })
      }
    }

    pc.oniceconnectionstatechange = () => {
      if (pc.iceConnectionState === 'disconnected' || pc.iceConnectionState === 'failed') {
        cleanupPeer(peerId)
      }
    }

    const peerData: RTCPeerData = {
      id: peerId,
      connection: pc,
      channel: dataChannel,
      status: 'connecting',
    }

    peersRef.current.set(peerId, peerData)
    setPeers(new Map(peersRef.current))

    return pc
  }

  // ─── Offer / Answer ───────────────────────────────────────────────
  const handleOffer = async (msg: RTCMessage) => {
    const pc = createPeerConnection(msg.from_id)
    await pc.setRemoteDescription(JSON.parse(msg.sdp!))
    const answer = await pc.createAnswer()
    await pc.setLocalDescription(answer)

    sendMessage({
      type: 'answer',
      from_id: config.peerId || 'client',
      to_id: msg.from_id,
      sdp: JSON.stringify(answer),
      timestamp: Date.now(),
    })
  }

  const handleAnswer = async (msg: RTCMessage) => {
    const peer = peersRef.current.get(msg.from_id)
    if (peer) {
      await peer.connection.setRemoteDescription(JSON.parse(msg.sdp!))
    }
  }

  const handleICECandidate = async (msg: RTCMessage) => {
    const peer = peersRef.current.get(msg.from_id)
    if (peer && msg.candidate) {
      try {
        await peer.connection.addIceCandidate(JSON.parse(msg.candidate))
      } catch (e) {
        console.error('[WebRTC] Error adding ICE candidate:', e)
      }
    }
  }

  // ─── Initiate Connection ──────────────────────────────────────────
  const connectToPeer = async (peerId: string) => {
    if (peersRef.current.has(peerId)) return

    const pc = createPeerConnection(peerId)
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)

    sendMessage({
      type: 'offer',
      from_id: config.peerId || 'client',
      to_id: peerId,
      sdp: JSON.stringify(offer),
      timestamp: Date.now(),
    })
  }

  // ─── Send Data ────────────────────────────────────────────────────
  const sendData = (peerId: string, data: string) => {
    const peer = peersRef.current.get(peerId)
    if (peer?.channel?.readyState === 'open') {
      peer.channel.send(data)
    }
  }

  const broadcastData = (data: string) => {
    peersRef.current.forEach((peer) => {
      if (peer.channel?.readyState === 'open') {
        peer.channel.send(data)
      }
    })
  }

  // ─── Send Signaling Message ───────────────────────────────────────
  const sendMessage = (msg: RTCMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg))
    }
  }

  // ─── Cleanup ──────────────────────────────────────────────────────
  const updatePeerState = (peerId: string, status: RTCPeerData['status']) => {
    const peer = peersRef.current.get(peerId)
    if (peer) {
      peer.status = status
      setPeers(new Map(peersRef.current))
    }
  }

  const cleanupPeer = (peerId: string) => {
    const peer = peersRef.current.get(peerId)
    if (peer) {
      peer.channel?.close()
      peer.connection.close()
      peersRef.current.delete(peerId)
      setPeers(new Map(peersRef.current))
      config.onPeerDisconnected?.(peerId)
    }
  }

  const disconnectAll = () => {
    peersRef.current.forEach((_, id) => cleanupPeer(id))
    wsRef.current?.close()
    setConnected(false)
  }

  // ─── Init ─────────────────────────────────────────────────────────
  useEffect(() => {
    connectSignaling()
    return () => disconnectAll()
  }, [])

  return {
    peers: Array.from(peers.values()),
    connected,
    error,
    connectToPeer,
    disconnectAll,
    sendData,
    broadcastData,
  }
}
