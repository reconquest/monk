package main

const SignaturePresence = "presence"

type PacketPresence struct {
	ID          string      `json:"id"`
	Fingerprint Fingerprint `json:"fingerprint"`
}

func (packet PacketPresence) Signature() string {
	return SignaturePresence
}

const SignatureError = "error"

type PacketError struct {
	Error string `json:"error"`
}

func (packet PacketError) Signature() string {
	return "error"
}

const SignatureQueryPeers = "peers/query"

type PacketQueryPeers struct {
}

func (packet PacketQueryPeers) Signature() string {
	return SignatureQueryPeers
}

const SignaturePeers = "peers"

type PacketPeers []Peer

func (packet PacketPeers) Signature() string {
	return SignaturePeers
}

const SignaturePeer = "peer"

type PacketPeer struct{ *Peer }

func (packet PacketPeer) Signature() string {
	return SignaturePeer
}

const SignatureTrustPeer = "peer/trust"

type PacketTrustPeer struct {
	ID string `json:"id"`
}

func (packet PacketTrustPeer) Signature() string {
	return SignatureTrustPeer
}

const SignatureQueryCertificate = "certificate/query"

type PacketQueryCertificate struct {
}

func (packet PacketQueryCertificate) Signature() string {
	return SignatureQueryCertificate
}

const SignatureCertificate = "certificate"

type PacketCertificate struct {
	Data []byte `json:"data"`
}

func (packet PacketCertificate) Signature() string {
	return SignatureCertificate
}

const SignatureQueryStream = "stream/query"

type PacketQueryStream struct {
	ID string `json:"id"`
}

func (packet PacketQueryStream) Signature() string {
	return SignatureQueryStream
}

const SignatureStream = "stream"

type PacketStream struct {
	Pipe  string `json:"pipe"`
	Buddy Peer   `json:"buddy"`
}

func (packet PacketStream) Signature() string {
	return SignatureStream
}

const SignatureEncryptConnection = "connection/encrypt"

type PacketEncryptConnection struct {
	ID          string      `json:"id"`
	Fingerprint Fingerprint `json:"fingerprint"`
}

func (packet PacketEncryptConnection) Signature() string {
	return SignatureEncryptConnection
}
