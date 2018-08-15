package main

func (monk *Monk) HandleQueryCertificate(Packet) (Packetable, error) {
	return PacketCertificate{
		Data: monk.security.CertificatePEM,
	}, nil
}
